package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	go src.waitForEnd(context.Background())

	return
}

// if it returns with an error, the APIKey field will be set to an empty string
func (s *YoutubeSource) GetAPIKey() error {

	APIFile, err := os.Open(os.Getenv("YOUTUBE_TOKEN_PATH"))
	if err != nil {
		return err
	}

	APIBytes, err := io.ReadAll(APIFile)
	if err != nil {
		return err
	}

	var token struct {
		Api string `json:"api"`
	}

	json.Unmarshal(APIBytes, &token)

	s.APIKey = token.Api

	err = s.CheckAPIKey()
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

func (s *YoutubeSource) Play(musicID MusicID) error {
	url, err := getAudioURL(musicID)
	if err != nil {
		return err
	}

	_, err = s.instance.Call("loadfile", url)
	if err == nil {
		s.isActive = true
	}
	return err
}

func (s *YoutubeSource) Search(query string) []MusicID {
	if s.APIKey == "" {
		return []MusicID{}
	}
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/search?part=snippet&key="+s.APIKey+"&type=video&q="+url.QueryEscape(query), nil)
	if err != nil {
		return []MusicID{}
	}

	res, err := client.Do(req)
	if err != nil {
		return []MusicID{}
	}

	all, err := io.ReadAll(res.Body)
	if err != nil {
		return []MusicID{}
	}

	defer res.Body.Close()

	var results YoutubeResponse
	json.Unmarshal(all, &results)

	ret := make([]MusicID, 0)
	for _, song := range results.Items {
		ret = append(ret, MusicID{
			TrackID:    song.ID.VideoID,
			SourceName: "youtube",
			Title:      song.Snippet.Title,
		})
	}

	return ret
}

func (s *YoutubeSource) ResolveTitle(musicID *MusicID) (string, error) {
	if musicID.Title != "" {
		return musicID.Title, nil
	}

	video, err := getVideo(musicID)
	if err != nil {
		return "", err
	}

	return video.Title, nil
}

func (c *YoutubeSource) BelongsToThis(query string) (bool, MusicID) {
	switch {
	case strings.HasPrefix(query, "https://www.youtube.com/watch?v="):
		m := MusicID{
			TrackID:    query[len("https://www.youtube.com/watch?v=") : len("https://www.youtube.com/watch?v=")+11],
			SourceName: "youtube",
		}
		m.Title, _ = c.ResolveTitle(&m)
		return true, m
	case strings.HasPrefix(query, "https://youtu.be/"):
		m := MusicID{
			TrackID:    query[len("https://youtu.be/") : len("https://youtu.be/")+11],
			SourceName: "youtube",
		}
		m.Title, _ = c.ResolveTitle(&m)
		return true, m
	case isYoutubeID(query):
		m := MusicID{
			TrackID:    query,
			SourceName: "youtube",
		}
		m.Title, _ = c.ResolveTitle(&m)
		return true, m
	default:
		return false, MusicID{}
	}
}

func getBestAudio(formats youtube.FormatList) youtube.Format {
	for _, format := range formats {
		if format.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
			return format
		}
	}
	return formats[0]
}

func getVideo(m *MusicID) (*youtube.Video, error) {
	id := m.youtube()
	yt := &youtube.Client{}
	return yt.GetVideo(id)
}

func getAudioURL(musicID MusicID) (string, error) {
	video, err := getVideo(&musicID)
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
