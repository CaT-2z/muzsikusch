package queue

import (
	"crypto/sha256"
	"encoding/base64"
	"math/rand"
	"time"
)

// TODO: ADD ENDPOINTS!!!
type Queue struct {
	Entries []Entry
}

// Adds unique ID to track, so only that one gets deleted
type Entry struct {
	MusicID
	// ID of the playlist its a part of
	Playlist   string
	UID        string
	PlaylistID string
	// In case the track starts from a custom timestamp or something
	StartTime float32
}

// TODO: Make a separate file for this
type Playlist struct {
	ID           string
	Title        string
	Creator      string
	ArtworkURL   string
	Date         time.Time
	ModifiedDate time.Time
	Tracks       []MusicID
	// If copied from a pre-existing playlist, this is a URL linking to that playlist e.g. Marcsello
	OriginalURL string
}

func NewQueue() (queue *Queue) {
	queue = &Queue{
		Entries: make([]Entry, 0),
	}
	return
}

// This copies
func (q *Queue) Append(track MusicID) Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}

	q.Entries = append(q.Entries, e)
	return e
}

func (q *Queue) AppendWithTime(track MusicID, timeStamp float32) Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := Entry{
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
	nq := make([]Entry, 0)
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
func (q *Queue) Push(track MusicID) Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}
	if len(q.Entries) < 1 {
		q.Entries = append([]Entry{q.Entries[0], e}, q.Entries[1:]...)
	} else {
		q.Entries = append(q.Entries, e)
	}
	return e
}

func (q *Queue) ForcePush(track MusicID, timeStamp float32) Entry {
	hash := sha256.Sum256([]byte(string(rand.Int31())))
	e := Entry{
		MusicID: track,
		UID:     base64.StdEncoding.EncodeToString(hash[:]),
	}
	if len(q.Entries) < 0 {
		q.Entries[0].StartTime = timeStamp
	}
	q.Entries = append([]Entry{e}, q.Entries...)
	return e
}

// Returns with the playlistID
func (q *Queue) AddMultiple(tracks []MusicID) string {
	playlistHash := sha256.Sum256([]byte(string(rand.Int31())))
	for _, track := range tracks {
		hash := sha256.Sum256([]byte(string(rand.Int31())))
		e := Entry{
			MusicID:    track,
			UID:        base64.StdEncoding.EncodeToString(hash[:]),
			PlaylistID: string(playlistHash[:]),
		}
		q.Entries = append(q.Entries, e)
	}
	return string(playlistHash[:])
}

// Returns with the playlistID
func (q *Queue) AddPlaylist(p Playlist) string {
	playlistHash := sha256.Sum256([]byte(string(rand.Int31())))
	for _, track := range p.Tracks {
		hash := sha256.Sum256([]byte(string(rand.Int31())))
		e := Entry{
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

func (q *Queue) Pop() (e Entry) {
	if len(q.Entries) < 2 {
		q.Entries = []Entry{}
		return Entry{}
	}
	q.Entries = q.Entries[1:]
	return q.Entries[0]
}

func (q *Queue) GetQueue() []Entry {
	if len(q.Entries) < 2 {
		return []Entry{}
	}
	return q.Entries[1:]
}

func (q *Queue) Flush() {
	q.Entries = make([]Entry, 0)
}

func (q *Queue) CurrentTrack() Entry {
	if len(q.Entries) == 0 {
		return Entry{}
	}
	return q.Entries[0]
}
