package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"muzsikusch/src/queue/entry"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kkdai/youtube/v2"
)

type YoutubeSource struct {
	MpvSource
	APIKey string
}

func NewYoutubeSource() (src *YoutubeSource, name string, err error) {

	name = "youtube"

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("NewYoutubeSource: Unable to start youtube service: %s", e)
		}
	}()

	mpv, err := NewMpvSource("/tmp/mpvsocket", context.Background())

	if err != nil {
		panic(err)
	}

	src = &YoutubeSource{
		MpvSource: *mpv,
	}

	otherErr := src.GetAPIKey()
	if otherErr != nil {
		fmt.Println("Search via youtube disabled, reason:", otherErr)
	}

	//go src.waitForEnd(context.Background())

	return
}

// if it returns with an error, the APIKey field will be set to an empty string
func (s *YoutubeSource) GetAPIKey() error {

	s.APIKey = os.Getenv("YOUTUBE_TOKEN")

	err := s.CheckAPIKey()
	if err != nil {
		s.APIKey = ""
		return err
	}

	return err
}

func (s *YoutubeSource) CheckAPIKey() (err error) {
	client := &http.Client{}

	//Is there a better request to test this with?
	req, err := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/search?part=snippet&key="+s.APIKey+"&type=video&q=cats", nil)
	if err != nil {
		return
	}

	res, err := client.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("Couldn't verify APIKey, response code: %d", res.StatusCode)
	}

	return
}

func (s *YoutubeSource) Play(MusicID entry.MusicID) error {
	url, err := getAudioURL(MusicID)
	if err != nil {
		return err
	}

	return s.PlayUrl(url)
}

func (s *YoutubeSource) Search(query string) []entry.MusicID {
	if s.APIKey == "" {
		return []entry.MusicID{}
	}
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/search?part=snippet&key="+s.APIKey+"&type=video&q="+url.QueryEscape(query), nil)
	if err != nil {
		return []entry.MusicID{}
	}

	res, err := client.Do(req)
	if err != nil {
		return []entry.MusicID{}
	}

	all, err := io.ReadAll(res.Body)
	if err != nil {
		return []entry.MusicID{}
	}

	defer res.Body.Close()

	var results YoutubeResponse
	json.Unmarshal(all, &results)

	ret := make([]entry.MusicID, 0)
	for _, song := range results.Items {
		ret = append(ret, s.getMusicIDFromCanonical(song.ID.VideoID))
	}

	return ret
}

func (s *YoutubeSource) ResolveTitle(MusicID *entry.MusicID) (string, error) {
	if MusicID.Title != "" {
		return MusicID.Title, nil
	}

	video, err := getVideo(MusicID)
	if err != nil {
		return "", err
	}

	return video.Title, nil
}

func (c *YoutubeSource) BelongsToThis(query string) (bool, entry.MusicID) {
	switch {
	case strings.HasPrefix(query, "https://www.youtube.com/watch?v="):
		m := c.getMusicIDFromCanonical(query[len("https://www.youtube.com/watch?v=") : len("https://www.youtube.com/watch?v=")+11])
		return true, m
	case strings.HasPrefix(query, "https://youtu.be/"):
		m := c.getMusicIDFromCanonical(query[len("https://youtu.be/") : len("https://youtu.be/")+11])
		return true, m
	case isYoutubeID(query):
		m := c.getMusicIDFromCanonical(query)
		return true, m
	default:
		return false, entry.MusicID{}
	}
}

func (c *YoutubeSource) getMusicIDFromCanonical(ID string) entry.MusicID {
	yt := &youtube.Client{}
	vid, _ := yt.GetVideo(ID)

	m := entry.MusicID{
		TrackID:    ID,
		SourceName: "youtube",
		Title:      vid.Title,
		Author:     vid.Author,
		ArtworkURL: vid.Thumbnails[0].URL,
		Duration:   vid.Duration,
	}
	return m
}

func getBestAudio(formats youtube.FormatList) youtube.Format {
	for _, format := range formats {
		if format.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
			return format
		}
	}
	return formats[0]
}

func getVideo(m *entry.MusicID) (*youtube.Video, error) {
	id := youtubeID(*m)
	yt := &youtube.Client{}
	return yt.GetVideo(id)
}

func getAudioURL(MusicID entry.MusicID) (string, error) {
	video, err := getVideo(&MusicID)
	if err != nil {
		return "", err
	}

	formats := video.Formats.WithAudioChannels()
	formats.Sort()

	best := getBestAudio(formats)

	yt := &youtube.Client{}
	//Only get audio stream
	url, err := yt.GetStreamURL(video, &best)

	if err != nil {
		return "", err
	}

	return url, nil
}

func isYoutubeID(query string) bool {
	if len(query) != 11 {
		return false
	}

	for _, c := range query {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_", c) {
			return false
		}
	}
	return true
}

func youtubeID(m entry.MusicID) string {
	if m.TrackID == "" {
		panic("Attempted to call youtube() on a MusicID without a youtube ID")
	}
	return m.TrackID
}
