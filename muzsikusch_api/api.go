package main

import (
	"fmt"
	"log"
)

type Muzsikusch struct {
	currentSource Source
	queue         []MusicID
	spotifySource *SpotifySource
	youtubeSource *YoutubeSource
}

func (m *Muzsikusch) Play(music_ID MusicID) error {
	if music_ID.isSpotify() {
		m.currentSource = m.spotifySource
		return m.spotifySource.Play(music_ID)
	}
	if music_ID.isYoutube() {
		m.currentSource = m.youtubeSource
		return m.youtubeSource.Play(music_ID)
	}
	return nil
}

func (m *Muzsikusch) Enqueue(music_id MusicID) error {
	m.queue = append(m.queue, music_id)
	fmt.Printf("Queue add: %v\n", m.queue)
	if len(m.queue) == 1 {
		return m.Play(music_id)
	}

	return nil
}

func (m *Muzsikusch) Pause() error {
	return m.currentSource.Pause()
}
func (m *Muzsikusch) Stop() error {
	return m.currentSource.Stop()
}
func (m *Muzsikusch) Skip() error {
	m.currentSource.Skip()
	if len(m.queue) > 0 {
		m.queue = m.queue[1:]
	}
	if len(m.queue) > 0 {
		return m.Play(m.queue[0])
	}

	return nil
}
func (m *Muzsikusch) Resume() error {
	return m.currentSource.Resume()
}
func (m *Muzsikusch) Forward(amm int) error {
	return m.currentSource.Forward(amm)
}
func (m *Muzsikusch) Reverse(amm int) error {
	return m.currentSource.Reverse(amm)
}
func (m *Muzsikusch) SetVolume(vol int) error {
	return m.currentSource.SetVolume(vol)
}
func (m *Muzsikusch) GetVolume() (int, error) {
	return m.currentSource.GetVolume()
}
func (m *Muzsikusch) Mute() error {
	return m.currentSource.Mute()
}

func (m *Muzsikusch) RegisterSource(source Source) {

}

func (m *Muzsikusch) UnregisterSource(source Source) error {
	return nil

}

func (m *Muzsikusch) Search(query, source string) MusicID {
	switch source {
	case "spotify":
		return m.spotifySource.Search(query)
	case "youtube":
		return m.youtubeSource.Search(query)
	default:
		panic("Unknown source")
	}
}

func (m *Muzsikusch) OnPlaybackFinished() {
	m.queue = m.queue[1:]
	log.Printf("Queue: %v\n", m.queue)
	if len(m.queue) > 0 {
		m.Play(m.queue[0])

		//TODO: Set current player
		if m.queue[0].isSpotify() {
			m.currentSource = m.spotifySource
		} else if m.queue[0].isYoutube() {
			m.currentSource = m.youtubeSource
		}
	}
}
