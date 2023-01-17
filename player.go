package main

import (
	"fmt"
	"log"
)

type Muzsikusch struct {
	currentSource Source
	queue         []MusicID
	sources       map[string]Source
	resolvers     map[string]TitleResolver
}

func NewMuzsikusch() *Muzsikusch {
	return &Muzsikusch{
		sources:   make(map[string]Source),
		resolvers: make(map[string]TitleResolver),
	}
}

func (m *Muzsikusch) Play(music_ID MusicID) error {
	src, ok := m.sources[music_ID.SourceName]
	if !ok {
		return fmt.Errorf("Source %s not registered", music_ID.SourceName)
	}

	m.currentSource = src

	return src.Play(music_ID)
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
	m.queue = m.queue[:0]
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

func (m *Muzsikusch) RegisterSource(source Source, name string) {
	m.sources[name] = source
}

func (m *Muzsikusch) UnregisterSource(name string) {
	m.sources[name] = nil
}

func (m *Muzsikusch) RegisterResolver(resolver TitleResolver, name string) {
	m.resolvers[name] = resolver
}

func (m *Muzsikusch) UnregisterResolver(name string) {
	m.resolvers[name] = nil
}

// TODO: Change this too
func (m *Muzsikusch) Search(query string) []MusicID {

	results := make([]MusicID, 0)
	for _, source := range m.sources {
		results = append(results, source.Search(query)...)
	}

	return results
}

func (m *Muzsikusch) OnPlaybackFinished() {
	m.queue = m.queue[1:]
	log.Printf("Queue: %v\n", m.queue)
	if len(m.queue) > 0 {
		m.Play(m.queue[0])

		// Set current player
		m.currentSource = m.sources[m.queue[0].SourceName]
	}
}

func (m *Muzsikusch) ResolveTitle(music_id *MusicID) (string, error) {
	resolver, ok := m.resolvers[music_id.SourceName]
	if !ok {
		log.Fatalf("Resolver %s not registered", music_id.SourceName)
	}

	return resolver.ResolveTitle(music_id)
}

func (m *Muzsikusch) GetQueue() []MusicID {
	return m.queue
}

// Registers the source if it went down without errors
func (m *Muzsikusch) SetupSource(source interface {
	Source
	TitleResolver
}, name string, err error) {
	if err != nil {
		fmt.Printf("SetupSource: Will start without %s: %s\n", name, err)
		return
	}
	m.RegisterSource(source, name)
	m.RegisterResolver(source, name)
	source.Register(m.OnPlaybackFinished)
}
