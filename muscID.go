package main

import (
	"github.com/zmb3/spotify/v2"
)

type MusicID struct {
	trackID    string `json:"-"`
	SourceName string `json:"-"`
	Title      string
}

// searchSource is always "spotify", general search wont work if spotify doesnt work TODO: change that
// TODO: Consider adding search to youtube too.
// TODO: Drop down bar for the search
func FromUser(query string, player *Muzsikusch, searchSource string) MusicID {
	if player == nil {
		panic("Attempted to search for a query without a searcher")
	}

	for _, source := range player.sources {
		if ok, mid := source.BelongsToThis(query); ok {
			return mid
		}
	}
	return player.Search(query, searchSource)
}

func (m MusicID) spotify() spotify.URI {
	if m.trackID == "" {
		panic("Attempted to call spotify() on a MusicID without a spotify URI")
	}
	return spotify.URI(m.trackID)
}

func (m MusicID) youtube() string {
	if m.trackID == "" {
		panic("Attempted to call youtube() on a MusicID without a youtube ID")
	}
	return m.trackID
}

func (m MusicID) isYoutube() bool {
	return m.SourceName == "youtube" && m.trackID != ""
}

func (m MusicID) isSpotify() bool {
	return m.SourceName == "spotify" && m.trackID != ""
}
