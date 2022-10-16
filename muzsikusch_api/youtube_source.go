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

func getBestAudio(formats youtube.FormatList) youtube.Format {
	for _, format := range formats {
		if format.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
			return format
		}
	}
	return formats[0]
}

func getAudioURL(musicID MusicID) (string, error) {
	id := musicID.youtube()
	yt := &youtube.Client{}
	video, err := yt.GetVideo(id)
	if err != nil {
		return "", err
	}

	formats := video.Formats.WithAudioChannels()
	formats.Sort()

	best := getBestAudio(formats)

	//Only get audio stream
	url, err := yt.GetStreamURL(video, &best)

	if err != nil {
		return "", err
	}

	return url, nil
}
