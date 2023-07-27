package queue

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	entry2 "muzsikusch/src/queue/entry"
	websocket2 "muzsikusch/src/websocket"
)

// TODO: ADD ENDPOINTS!!!
type Queue struct {
	Entries   []entry2.Entry
	wsmanager *websocket2.Manager
}

// Adds unique ID to track, so only that one gets deleted

func NewQueue() (queue *Queue) {
	queue = &Queue{
		Entries:   make([]entry2.Entry, 0),
		wsmanager: websocket2.NewManager(),
	}
	return
}

func (q *Queue) SetWSManager(man *websocket2.Manager) {
	q.wsmanager = man
}

// This copies
func (q *Queue) Append(track entry2.MusicID) entry2.Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := entry2.Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}

	q.Entries = append(q.Entries, e)
	fmt.Println("ADDED " + e.Title) //LOGGING
	q.wsmanager.WriteAll(websocket2.CreateAppendEvent(e))
	return e
}

func (q *Queue) AppendWithTime(track entry2.MusicID, timeStamp float32) entry2.Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := entry2.Entry{
		MusicID:   track,
		UID:       base64.StdEncoding.EncodeToString(hash[:]),
		StartTime: timeStamp,
	}

	q.Entries = append(q.Entries, e)
	q.wsmanager.WriteAll(websocket2.CreateAppendEvent(e))
	return e
}

// Boolean means the removal was succesful, the array changed
func (q *Queue) RemoveTrack(UID string) bool {
	for i, e := range q.Entries {
		if e.UID == UID {
			q.Entries = append(q.Entries[:i], q.Entries[i+1:]...)
			q.wsmanager.WriteAll(websocket2.CreateRemoveEvent(UID))
			return true
		}
	}
	return false
}

// Removes tracks by the playlistID
func (q *Queue) RemoveMultiple(UID string) (b bool) {
	nq := make([]entry2.Entry, 0)
	for _, e := range q.Entries {
		if e.PlaylistID != UID {
			nq = append(nq, e)
		} else {
			b = true
		}
	}
	q.Entries = nq
	q.wsmanager.WriteAll(websocket2.CreateRemoveEvent(UID))
	return
}

// Pushes the track to the front of the queue
func (q *Queue) Push(track entry2.MusicID) entry2.Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := entry2.Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}
	if len(q.Entries) < 1 {
		q.Entries = append(q.Entries, e)
	} else {
		q.Entries = append([]entry2.Entry{q.Entries[0], e}, q.Entries[1:]...)
	}
	q.wsmanager.WriteAll(websocket2.CreatePushEvent(e))
	return e
}

func (q *Queue) ForcePush(track entry2.MusicID, timeStamp float32) entry2.Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := entry2.Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}
	if len(q.Entries) < 0 {
		q.Entries[0].StartTime = timeStamp
	}
	q.Entries = append([]entry2.Entry{e}, q.Entries...)
	q.wsmanager.WriteAll(websocket2.CreatePushEvent(e))
	return e
}

// Returns with the playlistID
func (q *Queue) AddMultiple(tracks []entry2.MusicID) string {
	playlistHash := sha256.Sum256([]byte(string(rand.Int31())))
	for _, track := range tracks {
		hash := sha256.Sum256([]byte(string(rand.Int31())))
		e := entry2.Entry{
			MusicID:    track,
			UID:        base64.StdEncoding.EncodeToString(hash[:]),
			PlaylistID: string(playlistHash[:]),
		}
		q.Entries = append(q.Entries, e)
		q.wsmanager.WriteAll(websocket2.CreateAppendEvent(e))
	}
	return string(playlistHash[:])
}

// Returns with the playlistID
func (q *Queue) AddPlaylist(p entry2.Playlist) string {
	playlistHash := sha256.Sum256([]byte(string(rand.Int31())))
	for _, track := range p.Tracks {
		hash := sha256.Sum256([]byte(string(rand.Int31())))
		e := entry2.Entry{
			MusicID:    track,
			UID:        base64.StdEncoding.EncodeToString(hash[:]),
			PlaylistID: string(playlistHash[:]),
			Playlist:   p.ID,
		}
		q.Entries = append(q.Entries, e)
		q.wsmanager.WriteAll(websocket2.CreateAppendEvent(e))
	}
	return string(playlistHash[:])
}

func (q *Queue) Length() int {
	return len(q.Entries) - 1
}

// Ill need to look at this
func (q *Queue) Pop() (e entry2.Entry) {
	if len(q.Entries) < 1 {
		q.Entries = []entry2.Entry{}
		return entry2.Entry{}
	}
	e = q.Entries[0]
	q.Entries = q.Entries[1:]
	q.wsmanager.WriteAll(websocket2.CreateRemoveEvent(e.UID))
	return
}

func (q *Queue) GetQueue() []entry2.Entry {
	if len(q.Entries) < 2 {
		return []entry2.Entry{}
	}
	return q.Entries[1:]
}

func (q *Queue) Flush() {
	q.Entries = make([]entry2.Entry, 0)
}

func (q *Queue) CurrentTrack() entry2.Entry {
	if len(q.Entries) == 0 {
		return entry2.Entry{}
	}
	return q.Entries[0]
}
