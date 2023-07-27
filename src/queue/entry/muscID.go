package entry

import (
	"time"
)

/*
TODO: Implement:
  - ArtworkURL
  - Author
  - Duration
*/
type MusicID struct {
	ArtworkURL string        `json:"ArtworkURL"`
	TrackID    string        `json:"TrackID"`
	SourceName string        `json:"sourceName"`
	Title      string        `json:"title"`
	Author     string        `json:"Author"`
	Duration   time.Duration `json:"Duration"`
}

func (m MusicID) isYoutube() bool {
	return m.SourceName == "youtube" && m.TrackID != ""
}

func (m MusicID) isSpotify() bool {
	return m.SourceName == "spotify" && m.TrackID != ""
}
