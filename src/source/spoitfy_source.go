package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"muzsikusch/src/queue/entry"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// Source
type SpotifySource struct {
	client             *spotify.Client
	playerDevice       spotify.ID
	ctx                context.Context
	old_volume         int
	onPlaybackFinished func()
	dbusConn           *dbus.Conn
	waiterEnder        context.CancelFunc
}

func NewSpotifyFromToken(tokenPath string) (src *SpotifySource, name string, err error) {

	name = "spotify"

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("NewSpotifyFromToken: Unable to start Spotify service: %s", e)
		}
	}()

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	tok, err := getSpotifyToken(tokenPath)
	if err != nil {
		panic(err)
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(os.Getenv("REDIRECTURI")),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
		),
	)

	src = &SpotifySource{
		client:   spotify.New(auth.Client(context.Background(), &tok)),
		ctx:      context.Background(),
		dbusConn: conn,
	}

	return
}

func NewSpotifyWithAuth() *SpotifySource {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}

	src := &SpotifySource{
		ctx:      context.Background(),
		dbusConn: conn,
	}

	//Preform auth
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(os.Getenv("REDIRECTURI")),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
		),
	)

	//TODO: Randomize state
	state := "ABC123"
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	//Set up listener

	client := waitForAuth(state, auth)

	src.client = client

	return src
}

func (c *SpotifySource) Register(onPBFinished func()) {
	c.onPlaybackFinished = onPBFinished
}

func (c *SpotifySource) Play(music_id entry.MusicID) error {
	log.Printf("Playing %v\n", spotifyURI(music_id))
	uri := spotifyURI(music_id)
	opt := spotify.PlayOptions{
		DeviceID: &c.playerDevice,
		URIs:     []spotify.URI{uri},
	}

	if err := c.client.PlayOpt(c.ctx, &opt); err != nil {
		spotErr, ok := err.(spotify.Error)
		if !ok {
			return err
		}

		if spotErr.Status == 404 {
			//Device not found
			fmt.Println("Discovering devices")
			c.discoverDevices()
			return c.Play(music_id)
		}
	}

	//Check if we have a waiting ctx
	if c.waiterEnder != nil {
		c.waiterEnder()
	}

	ctx, waiterEnder := context.WithCancel(c.ctx)
	c.waiterEnder = waiterEnder

	go c.waitForEnd(ctx)
	return nil
}

func (c *SpotifySource) Pause() error {
	return c.client.Pause(c.ctx)
}

func (c *SpotifySource) Stop() error {
	c.waiterEnder()
	return c.client.Pause(c.ctx)
}
func (c *SpotifySource) Skip() error {
	c.waiterEnder()
	return c.client.Next(c.ctx)
}
func (c *SpotifySource) Resume() error {
	return c.client.Play(c.ctx)
}
func (c *SpotifySource) GetTimePos() (float32, error) {
	current, err := c.client.PlayerCurrentlyPlaying(c.ctx)
	if err != nil {
		return 0, err
	}
	x := float32(current.Progress) / 1000
	return x, nil
}

func (c *SpotifySource) Forward(ammount int) error {
	state, err := c.client.PlayerCurrentlyPlaying(c.ctx)
	if err != nil {
		return err
	}

	return c.client.Seek(c.ctx, state.Progress+ammount)
}
func (c *SpotifySource) Reverse(ammount int) error {
	return c.Forward(-ammount)
}
func (c *SpotifySource) SetVolume(vol int) error {
	return c.client.Volume(c.ctx, vol)
}
func (c *SpotifySource) GetVolume() (int, error) {
	state, err := c.client.PlayerState(c.ctx)

	//TODO: handle error
	if err != nil {
		return 0, err
	}

	return state.Device.Volume, nil
}

func (c *SpotifySource) Mute() error {
	vol, err := c.GetVolume()
	if err != nil {
		return err
	}

	if vol != 0 {
		c.old_volume = vol
		return c.SetVolume(0)
	} else {
		return c.SetVolume(c.old_volume)
	}
}

// I don't think you can specify the number of results in Spotify search
func (c *SpotifySource) Search(query string) []entry.MusicID {
	results, err := c.client.Search(c.ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found track %v\n", results.Tracks.Tracks[0].Name)

	tracks := make([]entry.MusicID, 0)

	for _, song := range results.Tracks.Tracks {
		tracks = append(tracks, entry.MusicID{
			TrackID:    string(song.URI),
			SourceName: "spotify",
			Title:      song.Name,
		})
	}

	return tracks
}

func (c *SpotifySource) ResolveTitle(mid *entry.MusicID) (string, error) {
	id := spotify.ID(spotifyURI(*mid)[14:])
	track, err := c.client.GetTrack(c.ctx, id)
	if err != nil {
		log.Printf("Error resolving title: %v\n", err)
		return "", err
	}

	return track.Name, nil
}

func (c *SpotifySource) discoverDevices() {
	devs, err := c.client.PlayerDevices(c.ctx)
	if err != nil {
		panic(err)
	}

	for dev := range devs {
		//TODO: Change this
		if devs[dev].Name == "archertwo" {
			c.playerDevice = devs[dev].ID
			return
		}
	}
}

func (c *SpotifySource) waitForEnd(ctx context.Context) {
	const WAIT_PERC = 0.95
	const WAIT_CUTOFF = 3 * time.Second

	obj := c.dbusConn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")

	done := ctx.Done()

	started := false

	for {
		//Check for cancelation
		select {
		case <-done:
			log.Println("Waiter canceled")
			return
		default:
		}

		//Get current position
		var pos uint64
		err := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.mpris.MediaPlayer2.Player", "Position").Store(&pos)
		if err != nil {
			panic(err)
		}

		//Get duration
		var metadata map[string]dbus.Variant
		err = obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.mpris.MediaPlayer2.Player", "Metadata").Store(&metadata)
		if err != nil {
			panic(err)
		}

		dur := metadata["mpris:length"].Value().(uint64)

		rem := time.Duration(dur-pos) * time.Microsecond

		//Wait
		if pos == 0 && started {
			c.onPlaybackFinished()
			return
		} else if rem < WAIT_CUTOFF && rem > 300*time.Millisecond {
			//Check every 0.1 seconds
			started = true
			time.Sleep(100 * time.Millisecond)
		} else {
			started = true
			cap := 20 * time.Second
			want := time.Duration(WAIT_PERC * float64(rem))
			wait := math.Min(float64(cap.Nanoseconds()), float64(want.Nanoseconds()))
			time.Sleep(time.Duration(wait))
		}

	}
}

func (c *SpotifySource) SaveToken(tokenPath string) {
	tok, err := c.client.Token()
	if err != nil {
		log.Printf("Cannot save token: %v\n", err)
		return
	}

	js, err := json.Marshal(tok)
	if err != nil {
		log.Printf("Cannot save token: %v\n", err)
		return
	}

	err = os.WriteFile(tokenPath, js, 0600)
	if err != nil {
		log.Printf("Cannot save token: %v\n", err)
		return
	}

	log.Println("Saved token")
}

func (c *SpotifySource) BelongsToThis(query string) (bool, entry.MusicID) {
	switch {
	case strings.HasPrefix(query, "spotify:track:"):
		m := entry.MusicID{
			TrackID:    query,
			SourceName: "spotify",
		}
		m.Title, _ = c.ResolveTitle(&m)
		return true, m
	case strings.HasPrefix(query, "https://open.spotify.com/track/"):
		m := entry.MusicID{
			TrackID:    "spotify:track:" + query[len("https://open.spotify.com/track/"):],
			SourceName: "spotify",
		}
		m.Title, _ = c.ResolveTitle(&m)
		return true, m
	case isSpotifyID(query):
		m := entry.MusicID{
			TrackID:    "spotify:track:" + query,
			SourceName: "spotify",
		}
		m.Title, _ = c.ResolveTitle(&m)
		return true, m
	default:
		return false, entry.MusicID{}
	}
}

func getSpotifyToken(tokenPath string) (oauth2.Token, error) {
	var token oauth2.Token

	//Check if we have a token
	tokenFile, err := os.Open(tokenPath)
	if err != nil {
		return token, err
	}

	//Read token
	tokenBytes, err := io.ReadAll(tokenFile)
	if err != nil {
		return token, err
	}

	//Parse token
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return token, err
	}

	return token, nil
}

func isSpotifyID(query string) bool {
	if len(query) != 22 {
		return false
	}

	for _, c := range query {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", c) {
			return false
		}
	}
	return true
}

func spotifyURI(m entry.MusicID) spotify.URI {
	if m.TrackID == "" {
		panic("Attempted to call spotify() on a MusicID without a spotify URI")
	}
	return spotify.URI(m.TrackID)
}
