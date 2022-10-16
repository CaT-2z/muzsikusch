package main

import (
	"context"

	"github.com/kkdai/youtube/v2"
)

type YoutubeSource struct {
	MpvSource
}

func NewYoutubeSource() *YoutubeSource {
	mpv, err := NewMpvSource("/tmp/mpvsocket", context.Background())
	if err != nil {
		panic(err)
	}

	return &YoutubeSource{
		MpvSource: *mpv,
	}
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
