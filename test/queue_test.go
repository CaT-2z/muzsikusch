package test

import (
	"muzsikusch/queue"
	"muzsikusch/queue/entry"
	"testing"
)

func TestQueueEmpty(t *testing.T) {
	q := queue.NewQueue()
	if len(q.GetQueue()) != 0 {
		t.Errorf("Queue does`t start empty")
	}
}

func TestQueueAppend(t *testing.T) {
	q := queue.NewQueue()

	music := entry.MusicID{
		ArtworkURL: "abc",
		TrackID:    "abc",
		SourceName: "abc",
		Title:      "abc",
		Author:     "abc",
		Duration:   10000,
	}

	entry := q.Append(music)

	if entry.MusicID != music {
		t.Errorf("Appended track doesn't match")
	}

	current := q.CurrentTrack()

	if current.MusicID != music {
		t.Errorf("CurrentTrack doesn`t match")
	}

	entry = q.Pop()

	if entry.MusicID != music {
		t.Errorf("Popped track doesn't match")
	}

}

func TestPushForcePush(t *testing.T) {
	q := queue.NewQueue()

	m1 := entry.MusicID{
		ArtworkURL: "a",
		TrackID:    "a",
		SourceName: "a",
		Title:      "a",
		Author:     "a",
		Duration:   0,
	}

	m2 := entry.MusicID{
		ArtworkURL: "b",
		TrackID:    "b",
		SourceName: "b",
		Title:      "b",
		Author:     "b",
		Duration:   0,
	}

	m3 := entry.MusicID{
		ArtworkURL: "c",
		TrackID:    "c",
		SourceName: "c",
		Title:      "c",
		Author:     "c",
		Duration:   0,
	}

	q.Append(m1)
	q.Append(m1)
	q.Push(m2)

	if q.GetQueue()[0].MusicID != m2 {
		t.Errorf("Push doesnt push to start")
	}

	q.ForcePush(m3, 0)

	if q.CurrentTrack().MusicID != m3 {
		t.Errorf("ForcePush doesnt push to start")
	}
}
