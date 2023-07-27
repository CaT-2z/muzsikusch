package entry

import "time"

type Entry struct {
	MusicID
	// ID of the playlist its a part of
	Playlist   string
	UID        string
	PlaylistID string
	// In case the track starts from a custom timestamp or something
	StartTime float32
}

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
