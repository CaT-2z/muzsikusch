package queue

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	entry "muzsikusch/queue/entry"
	"muzsikusch/websocket"
)

// TODO: ADD ENDPOINTS!!!
type Queue struct {
	Entries   []entry.Entry
	wsmanager *websocket.Manager
}

// Adds unique ID to track, so only that one gets deleted

func NewQueue() (queue *Queue) {
	queue = &Queue{
		Entries: make([]entry.Entry, 0),
	}
	return
}

func (q *Queue) SetWSManager(man *websocket.Manager) {
	q.wsmanager = man
}

// This copies
func (q *Queue) Append(track entry.MusicID) entry.Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := entry.Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}

	q.Entries = append(q.Entries, e)
	fmt.Println("ADDED " + e.Title) //LOGGING
	return e
}

func (q *Queue) AppendWithTime(track entry.MusicID, timeStamp float32) entry.Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := entry.Entry{
		MusicID:   track,
		UID:       base64.StdEncoding.EncodeToString(hash[:]),
		StartTime: timeStamp,
	}

	q.Entries = append(q.Entries, e)
	return e
}

// Boolean means the removal was succesful, the array changed
func (q *Queue) RemoveTrack(UID string) bool {
	for i, e := range q.Entries {
		if e.UID == UID {
			q.Entries = append(q.Entries[:i], q.Entries[i+1:]...)
			return true
		}
	}
	return false
}

// Removes tracks by the playlistID
func (q *Queue) RemoveMultiple(UID string) (b bool) {
	nq := make([]entry.Entry, 0)
	for _, e := range q.Entries {
		if e.PlaylistID != UID {
			nq = append(nq, e)
		} else {
			b = true
		}
	}
	q.Entries = nq
	return
}

// Pushes the track to the front of the queue
func (q *Queue) Push(track entry.MusicID) entry.Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := entry.Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}
	if len(q.Entries) < 1 {
		q.Entries = append([]entry.Entry{q.Entries[0], e}, q.Entries[1:]...)
	} else {
		q.Entries = append(q.Entries, e)
	}
	return e
}

func (q *Queue) ForcePush(track entry.MusicID, timeStamp float32) entry.Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := entry.Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}
	if len(q.Entries) < 0 {
		q.Entries[0].StartTime = timeStamp
	}
	q.Entries = append([]entry.Entry{e}, q.Entries...)
	return e
}

// Returns with the playlistID
func (q *Queue) AddMultiple(tracks []entry.MusicID) string {
	playlistHash := sha256.Sum256([]byte(string(rand.Int31())))
	for _, track := range tracks {
		hash := sha256.Sum256([]byte(string(rand.Int31())))
		e := entry.Entry{
			MusicID:    track,
			UID:        base64.StdEncoding.EncodeToString(hash[:]),
			PlaylistID: string(playlistHash[:]),
		}
		q.Entries = append(q.Entries, e)
	}
	return string(playlistHash[:])
}

// Returns with the playlistID
func (q *Queue) AddPlaylist(p entry.Playlist) string {
	playlistHash := sha256.Sum256([]byte(string(rand.Int31())))
	for _, track := range p.Tracks {
		hash := sha256.Sum256([]byte(string(rand.Int31())))
		e := entry.Entry{
			MusicID:    track,
			UID:        base64.StdEncoding.EncodeToString(hash[:]),
			PlaylistID: string(playlistHash[:]),
			Playlist:   p.ID,
		}
		q.Entries = append(q.Entries, e)
	}
	return string(playlistHash[:])
}

func (q *Queue) Length() int {
	return len(q.Entries) - 1
}

func (q *Queue) Pop() (e entry.Entry) {
	if len(q.Entries) < 2 {
		q.Entries = []entry.Entry{}
		return entry.Entry{}
	}
	q.Entries = q.Entries[1:]
	return q.Entries[0]
}

func (q *Queue) GetQueue() []entry.Entry {
	if len(q.Entries) < 2 {
		return []entry.Entry{}
	}
	return q.Entries[1:]
}

func (q *Queue) Flush() {
	q.Entries = make([]entry.Entry, 0)
}

func (q *Queue) CurrentTrack() entry.Entry {
	if len(q.Entries) == 0 {
		return entry.Entry{}
	}
	return q.Entries[0]
}
