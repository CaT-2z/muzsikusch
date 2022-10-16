package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/zmb3/spotify/v2"
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

func (c *SpotifySource) Register(onPBFinished func()) {
	c.onPlaybackFinished = onPBFinished
}

func (c *SpotifySource) Play(music_id MusicID) error {
	log.Printf("Playing %v\n", music_id.spotify())
	uri := music_id.spotify()
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

func (c *SpotifySource) Search(query string) MusicID {
	results, err := c.client.Search(c.ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found track %v\n", results.Tracks.Tracks[0].Name)
	return MusicID{
		spotifyURI: string(results.Tracks.Tracks[0].URI),
		SourceName: "spotify",
		Title:      results.Tracks.Tracks[0].Name,
	}
}

func (c *SpotifySource) ResolveTitle(mid *MusicID) (string, error) {
	track, err := c.client.GetTrack(c.ctx, spotify.ID(mid.spotify()))
	if err != nil {
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
	log.Println("Waiter started")
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
		log.Printf("Remaining: %v\n", rem)

		//Wait
		if pos == 0 && started {
			log.Println("FINISHED POS==0")
			c.onPlaybackFinished()
			return
		} else if rem < WAIT_CUTOFF && rem > 300*time.Millisecond {
			//Check every 0.1 seconds
			started = true
			log.Println("Waiting for 100ms")
			time.Sleep(100 * time.Millisecond)
		} else {
			started = true
			cap := 20 * time.Second
			want := time.Duration(WAIT_PERC * float64(rem))
			wait := math.Min(float64(cap.Nanoseconds()), float64(want.Nanoseconds()))
			log.Printf("Waiting for %v\n", time.Duration(wait))
			time.Sleep(time.Duration(wait))
		}

	}
}
