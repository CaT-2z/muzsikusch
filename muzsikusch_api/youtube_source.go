package main

import (
	"github.com/dexterlb/mpvipc"
	"github.com/kkdai/youtube/v2"
)

type YoutubeSource struct {
	instance mpvipc.Connection
	events   chan *mpvipc.Event
	stopChan chan struct{}
}

func (s *YoutubeSource) play(musicID MusicID) error {
	url, err := getAudioURL(musicID)
	if err != nil {
		return err
	}

	_, err = s.instance.Call("loadfile", url)
	return err
}
func (s *YoutubeSource) stop() error {
	_, err := s.instance.Call("stop")
	return err
}
func (s *YoutubeSource) skip() error {
	_, err := s.instance.Call("playlist-next", "force")
	return err
}
func (s *YoutubeSource) pause() error {
	return s.instance.Set("pause", true)
}
func (s *YoutubeSource) resume() error {
	return s.instance.Set("pause", false)
}
func (s *YoutubeSource) forward(amm int) error {
	_, err := s.instance.Call("seek", amm)
	return err
}
func (s *YoutubeSource) reverse(amm int) error {
	return s.forward(-amm)
}
func (s *YoutubeSource) setVolume(vol int) error {
	return s.instance.Set("volume", vol)
}
func (s *YoutubeSource) getVolume() (int, error) {
	vol, err := s.instance.Get("volume")
	if err != nil {
		return 0, err
	}
	return vol.(int), nil
}
func (s *YoutubeSource) mute() error {
	_, err := s.instance.Call("cycle", "mute")
	return err
}

func (s *YoutubeSource) search(query string) MusicID {
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
