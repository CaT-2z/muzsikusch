package main

import (
	"github.com/zmb3/spotify/v2"
)

type MusicID struct {
	TrackID    string `json:"TrackID"`
	SourceName string `json:"sourceName"`
	Title      string `json:"title"`
}

// searchSource is always "spotify", general search wont work if spotify doesnt work TODO: change that
// TODO: Drop down bar for the search
func FromUser(query string, player *Muzsikusch) []MusicID {

	for _, source := range player.sources {
		if ok, mid := source.BelongsToThis(query); ok {
			return []MusicID{mid}
		}
	}
	if player == nil {
		panic("Attempted to search for a query without a searcher")
	}
	// I'm trying to remove this part
	return player.Search(query)
}

func (m MusicID) spotify() spotify.URI {
	if m.TrackID == "" {
		panic("Attempted to call spotify() on a MusicID without a spotify URI")
	}
	return spotify.URI(m.TrackID)
}

func (m MusicID) youtube() string {
	if m.TrackID == "" {
		panic("Attempted to call youtube() on a MusicID without a youtube ID")
	}
	return m.TrackID
}

func (m MusicID) isYoutube() bool {
	return m.SourceName == "youtube" && m.TrackID != ""
}

func (m MusicID) isSpotify() bool {
	return m.SourceName == "spotify" && m.TrackID != ""
}
