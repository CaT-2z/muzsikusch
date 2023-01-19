package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/kkdai/youtube/v2"
)

type YoutubeSource struct {
	MpvSource
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
	go src.waitForEnd(context.Background())

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

func (s *YoutubeSource) Search(query string) MusicID {
	panic("Cannot search youtube")
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
			trackID:    query[len("https://www.youtube.com/watch?v=") : len("https://www.youtube.com/watch?v=")+11],
			SourceName: "youtube",
		}
		m.Title, _ = c.ResolveTitle(&m)
		return true, m
	case strings.HasPrefix(query, "https://youtu.be/"):
		m := MusicID{
			trackID:    query[len("https://youtu.be/") : len("https://youtu.be/")+11],
			SourceName: "youtube",
		}
		m.Title, _ = c.ResolveTitle(&m)
		return true, m
	case isYoutubeID(query):
		m := MusicID{
			trackID:    query,
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
