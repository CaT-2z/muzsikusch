package main

import (
	"strings"

	"github.com/zmb3/spotify/v2"
)

type MusicID struct {
	SpotifyURI string
	YoutubeID  string
}

func FromUser(query string, searcher *Muzsikusch, searchSource string) MusicID {
	switch {
	case strings.HasPrefix(query, "spotify:track:"):
		return FromSpotifyID(query[len("spotify:track:"):])
	case strings.HasPrefix(query, "https://open.spotify.com/track/"):
		return FromSpotifyID(query[len("https://open.spotify.com/track/"):])
	case strings.HasPrefix(query, "https://www.youtube.com/watch?v="):
		return FromYoutubeID(query[len("https://www.youtube.com/watch?v="):])
	case strings.HasPrefix(query, "https://youtu.be/"):
		return FromYoutubeID(query[len("https://youtu.be/"):])
	case isSpotifyID(query):
		return FromSpotifyID(query)
	case isYoutubeID(query):
		return FromYoutubeID(query)
	default:
		//Search for the query
		if searcher == nil {
			panic("Attempted to search for a query without a searcher")
		}
		return searcher.Search(query, searchSource)
	}
}

func FromSpotifyID(id string) MusicID {
	return MusicID{
		SpotifyURI: "spotify:track:" + id[:22],
	}
}

func FromYoutubeID(id string) MusicID {
	return MusicID{
		YoutubeID: id[:11],
	}
}

func (m MusicID) spotify() spotify.URI {
	if m.SpotifyURI == "" {
		panic("Attempted to call spotify() on a MusicID without a spotify URI")
	}
	return spotify.URI(m.SpotifyURI)
}

func (m MusicID) youtube() string {
	if m.YoutubeID == "" {
		panic("Attempted to call youtube() on a MusicID without a youtube ID")
	}
	return m.YoutubeID
}

func (m MusicID) isYoutube() bool {
	return m.YoutubeID != ""
}

func (m MusicID) isSpotify() bool {
	return m.SpotifyURI != ""
}

func isSpotifyID(query string) bool {
	if len(query) != 22 {
		return false
	}

	for _, c := range query {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", c) {
			return false
		}
	}
	return true
}

func isYoutubeID(query string) bool {
	if len(query) != 11 {
		return false
	}

	for _, c := range query {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_", c) {
			return false
		}
	}
	return true
}
