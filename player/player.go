package player

import (
	"fmt"
	"log"
	"muzsikusch/queue"
	entry "muzsikusch/queue/entry"
	"muzsikusch/source"
	"muzsikusch/websocket"
)

type Muzsikusch struct {
	currentSource source.Source
	queue         queue.Queue
	sources       map[string]source.Source
	resolvers     map[string]source.TitleResolver
	wsmanager     *websocket.Manager
}

func NewMuzsikusch() *Muzsikusch {
	return &Muzsikusch{
		queue:     *queue.NewQueue(),
		sources:   make(map[string]source.Source),
		resolvers: make(map[string]source.TitleResolver),
	}
}

//NOTE: player sends playback events, queue sends queue events

// TODO: Im not sure this is a good way of doing it, but it will work for now
func (m *Muzsikusch) SetWSManager(man *websocket.Manager) {
	m.wsmanager = man
	m.queue.SetWSManager(man)
}

func (m *Muzsikusch) Play(ent entry.Entry) error {
	fmt.Println("STARTED PLAYING " + ent.Title) //LOGGING
	src, ok := m.sources[ent.SourceName]
	if !ok {
		return fmt.Errorf("Source %s not registered", ent.SourceName)
	}
	m.currentSource = src
	defer m.wsmanager.WriteAll(websocket.CreateTrackStartEvent(ent))
	return src.Play(ent.MusicID)
}

func (m *Muzsikusch) Enqueue(music_id entry.MusicID) error {
	var ent entry.Entry
	if (m.queue.CurrentTrack() == entry.Entry{}) {
		ent = m.queue.Append(music_id)
		fmt.Printf("Queue add: %v\n", m.queue)
		return m.Play(ent)
	}
	m.queue.Append(music_id)
	return nil
}

func (m *Muzsikusch) Push(music_id entry.MusicID) error {
	if (m.queue.CurrentTrack() == entry.Entry{}) {
		fmt.Printf("Queue add: %v\n", m.queue)
		return m.Play(m.queue.Push(music_id))
	}
	m.queue.Push(music_id)
	return nil
}

func (m *Muzsikusch) Pause() error {
	err := m.currentSource.Pause()
	if err == nil {
		x, err := m.currentSource.GetTimePos()
		if err == nil {
			m.wsmanager.WriteAll(websocket.CreatePauseEvent(x))
		}
	}
	return err
}
func (m *Muzsikusch) Stop() error {
	m.queue.Flush()
	return m.currentSource.Stop()
}
func (m *Muzsikusch) Skip() error {
	m.currentSource.Skip()
	if m.queue.Length() > 0 {
		m.Play(m.queue.Pop())
	}
	m.queue.Pop()
	return nil
}
func (m *Muzsikusch) Resume() error {
	err := m.currentSource.Resume()
	if err == nil {
		x, err := m.currentSource.GetTimePos()
		if err == nil {
			m.wsmanager.WriteAll(websocket.CreateUnpauseEvent(x))
		}
	}
	return err
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
func (m *Muzsikusch) Search(query string) []entry.MusicID {

	results := make([]entry.MusicID, 0)
	for _, src := range m.sources {
		results = append(results, src.Search(query)...)
	}

	return results
}

func (m *Muzsikusch) OnPlaybackFinished() {
	if m.queue.Length() > 0 {
		m.Play(m.queue.Pop())
	}
	m.queue.Pop()
	log.Printf("Queue: %v\n", m.queue)
}

func (m *Muzsikusch) ResolveTitle(music_id *entry.MusicID) (string, error) {
	resolver, ok := m.resolvers[music_id.SourceName]
	if !ok {
		log.Fatalf("Resolver %s not registered", music_id.SourceName)
	}

	return resolver.ResolveTitle(music_id)
}

func (m *Muzsikusch) GetQueue() []entry.MusicID {
	q := []entry.MusicID{}
	//TODO: THIS shouldnt be like this
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
func (player *Muzsikusch) FromUser(query string) []entry.MusicID {

	for _, source := range player.sources {
		if ok, mid := source.BelongsToThis(query); ok {
			return []entry.MusicID{mid}
		}
	}
	if player == nil {
		panic("Attempted to search for a query without a searcher")
	}
	// I'm trying to remove this part
	return player.Search(query)
}
