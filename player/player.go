package player

import (
	"fmt"
	"log"
	"muzsikusch/queue"
	"muzsikusch/source"
)

type Muzsikusch struct {
	currentSource source.Source
	queue         queue.Queue
	sources       map[string]source.Source
	resolvers     map[string]source.TitleResolver
}

func NewMuzsikusch() *Muzsikusch {
	return &Muzsikusch{
		queue:     *queue.NewQueue(),
		sources:   make(map[string]source.Source),
		resolvers: make(map[string]source.TitleResolver),
	}
}

func (m *Muzsikusch) Play(music_ID queue.MusicID) error {
	src, ok := m.sources[music_ID.SourceName]
	if !ok {
		return fmt.Errorf("Source %s not registered", music_ID.SourceName)
	}
	m.currentSource = src
	return src.Play(music_ID)
}

func (m *Muzsikusch) Enqueue(music_id queue.MusicID) error {
	if (m.queue.CurrentTrack() == queue.Entry{}) {
		fmt.Printf("Queue add: %v\n", m.queue)
		return m.Play(m.queue.Append(music_id).MusicID)
	}
	m.queue.Append(music_id)
	return nil
}

func (m *Muzsikusch) Pause() error {
	return m.currentSource.Pause()
}
func (m *Muzsikusch) Stop() error {
	m.queue.Flush()
	return m.currentSource.Stop()
}
func (m *Muzsikusch) Skip() error {
	m.currentSource.Skip()
	if m.queue.Length() > 0 {
		m.Play(m.queue.Pop().MusicID)
	}
	m.queue.Pop()
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

func (m *Muzsikusch) RegisterSource(source source.Source, name string) {
	m.sources[name] = source
}

func (m *Muzsikusch) UnregisterSource(name string) {
	m.sources[name] = nil
}

func (m *Muzsikusch) RegisterResolver(resolver source.TitleResolver, name string) {
	m.resolvers[name] = resolver
}

func (m *Muzsikusch) UnregisterResolver(name string) {
	m.resolvers[name] = nil
}

// TODO: Change this too
func (m *Muzsikusch) Search(query string) []queue.MusicID {

	results := make([]queue.MusicID, 0)
	for _, src := range m.sources {
		results = append(results, src.Search(query)...)
	}

	return results
}

func (m *Muzsikusch) OnPlaybackFinished() {
	if m.queue.Length() > 0 {
		m.Play(m.queue.Pop().MusicID)
	}
	m.queue.Pop()
	log.Printf("Queue: %v\n", m.queue)
}

func (m *Muzsikusch) ResolveTitle(music_id *queue.MusicID) (string, error) {
	resolver, ok := m.resolvers[music_id.SourceName]
	if !ok {
		log.Fatalf("Resolver %s not registered", music_id.SourceName)
	}

	return resolver.ResolveTitle(music_id)
}

func (m *Muzsikusch) GetQueue() []queue.MusicID {
	q := []queue.MusicID{}
	for _, e := range m.queue.Entries {
		q = append(q, e.MusicID)
	}
	return q
}

// Registers the source if it went down without errors
func (m *Muzsikusch) SetupSource(source interface {
	source.Source
	source.TitleResolver
}, name string, err error) {
	if err != nil {
		fmt.Printf("SetupSource: Will start without %s: %s\n", name, err)
		return
	}
	m.RegisterSource(source, name)
	m.RegisterResolver(source, name)
	source.Register(m.OnPlaybackFinished)
}

// searchSource is always "spotify", general search wont work if spotify doesnt work TODO: change that
// TODO: Drop down bar for the search
func (player *Muzsikusch) FromUser(query string) []queue.MusicID {

	for _, source := range player.sources {
		if ok, mid := source.BelongsToThis(query); ok {
			return []queue.MusicID{mid}
		}
	}
	if player == nil {
		panic("Attempted to search for a query without a searcher")
	}
	// I'm trying to remove this part
	return player.Search(query)
}
