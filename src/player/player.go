package player

import (
	"fmt"
	"github.com/jonhoo/go-events"
	"log"
	"muzsikusch/src/queue"
	entry2 "muzsikusch/src/queue/entry"
	source2 "muzsikusch/src/source"
	websocket2 "muzsikusch/src/websocket"
)

type Muzsikusch struct {
	currentSource source2.Source
	queue         queue.Queue
	sources       map[string]source2.Source
	resolvers     map[string]source2.TitleResolver
	wsmanager     *websocket2.Manager
}

//NOTE: player sends playback events, queue sends queue events

// TODO: Im not sure this is a good way of doing it, but it will work for now
func (m *Muzsikusch) SetWSManager(man *websocket2.Manager) {
	m.wsmanager = man
	m.queue.SetWSManager(man)
}

func (m *Muzsikusch) Play(ent entry2.Entry) error {
	fmt.Println("STARTED PLAYING " + ent.Title) //LOGGING
	src, ok := m.sources[ent.SourceName]
	if !ok {
		return fmt.Errorf("Source %s not registered", ent.SourceName)
	}
	m.currentSource = src
	defer m.wsmanager.WriteAll(websocket2.CreateTrackStartEvent(ent))
	return src.Play(ent.MusicID)
}

func (m *Muzsikusch) Enqueue(music_id entry2.MusicID) error {
	var ent entry2.Entry
	if (m.queue.CurrentTrack() == entry2.Entry{}) {
		ent = m.queue.Append(music_id)
		fmt.Printf("Queue add: %v\n", m.queue)
		return m.Play(ent)
	}
	m.queue.Append(music_id)
	return nil
}

func (m *Muzsikusch) Push(music_id entry2.MusicID) error {
	isFirst := (m.queue.CurrentTrack() == entry2.Entry{})
	ent := m.queue.Push(music_id)
	m.wsmanager.WriteAll(websocket2.CreatePushEvent(ent))
	if isFirst {
		fmt.Printf("Queue add: %v\n", m.queue)
		return m.Play(ent)
	}
	return nil
}

func (m *Muzsikusch) Pause() error {
	err := m.currentSource.Pause()
	if err == nil {
		x, err := m.currentSource.GetTimePos()
		if err == nil {
			m.wsmanager.WriteAll(websocket2.CreatePauseEvent(x))
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
	m.wsmanager.WriteAll(websocket2.CreateRemoveEvent(m.queue.Pop().UID))
	return nil
}

func (m *Muzsikusch) Remove(UID string) bool {
	m.wsmanager.WriteAll(websocket2.CreateRemoveEvent(UID))
	return m.queue.RemoveTrack(UID)
}

func (m *Muzsikusch) Resume() error {
	err := m.currentSource.Resume()
	if err == nil {
		x, err := m.currentSource.GetTimePos()
		if err == nil {
			m.wsmanager.WriteAll(websocket2.CreateUnpauseEvent(x))
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

func (m *Muzsikusch) RegisterSource(source source2.Source, name string) {
	m.sources[name] = source
}

func (m *Muzsikusch) UnregisterSource(name string) {
	m.sources[name] = nil
}

func (m *Muzsikusch) RegisterResolver(resolver source2.TitleResolver, name string) {
	m.resolvers[name] = resolver
}

func (m *Muzsikusch) UnregisterResolver(name string) {
	m.resolvers[name] = nil
}

// TODO: Change this too
func (m *Muzsikusch) Search(query string) []entry2.MusicID {

	results := make([]entry2.MusicID, 0)
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

func (m *Muzsikusch) ResolveTitle(music_id *entry2.MusicID) (string, error) {
	resolver, ok := m.resolvers[music_id.SourceName]
	if !ok {
		log.Fatalf("Resolver %s not registered", music_id.SourceName)
	}

	return resolver.ResolveTitle(music_id)
}

func (m *Muzsikusch) GetQueue() []entry2.MusicID {
	q := []entry2.MusicID{}
	//TODO: THIS shouldnt be like this
	for _, e := range m.queue.Entries {
		q = append(q, e.MusicID)
	}
	return q
}

// Registers the source if it went down without errors
func (m *Muzsikusch) SetupSource(source interface {
	source2.Source
	source2.TitleResolver
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
func (player *Muzsikusch) FromUser(query string) []entry2.MusicID {

	for _, source := range player.sources {
		if ok, mid := source.BelongsToThis(query); ok {
			return []entry2.MusicID{mid}
		}
	}
	if player == nil {
		panic("Attempted to search for a query without a searcher")
	}
	// I'm trying to remove this part
	return player.Search(query)
}

func NewMuzsikusch() *Muzsikusch {
	m := Muzsikusch{
		queue:     *queue.NewQueue(),
		sources:   make(map[string]source2.Source),
		resolvers: make(map[string]source2.TitleResolver),
	}

	//Creates even handlers for the events
	eventHandlers := map[string]func(interface{}){
		"pause":   func(interface{}) { m.Pause() },
		"unpause": func(interface{}) { m.Resume() },
	}

	//Registers these, starts them in a goroutine
	for ev := range eventHandlers {
		str := ev
		go func() {
			chn := events.Listen(str)
			defer close(chn)
			for e := range chn {
				eventHandlers[str](e.Data)
			}
		}()
	}

	return &m
}
