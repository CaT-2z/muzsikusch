package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/zmb3/spotify/v2"
)

type MusicID struct {
	spotifyURI string `json:"-"`
	YoutubeID  string `json:"-"`
	SourceName string `json:"-"`
	Title      string
}

func FromUser(query string, searcher *Muzsikusch, searchSource string, resolver TitleResolver) MusicID {
	switch {
	case strings.HasPrefix(query, "spotify:track:"):
		return FromSpotifyID(query[len("spotify:track:"):], resolver)
	case strings.HasPrefix(query, "https://open.spotify.com/track/"):
		return FromSpotifyID(query[len("https://open.spotify.com/track/"):], resolver)
	case strings.HasPrefix(query, "https://www.youtube.com/watch?v="):
		return FromYoutubeID(query[len("https://www.youtube.com/watch?v="):], resolver)
	case strings.HasPrefix(query, "https://youtu.be/"):
		return FromYoutubeID(query[len("https://youtu.be/"):], resolver)
	case isSpotifyID(query):
		return FromSpotifyID(query, resolver)
	case isYoutubeID(query):
		return FromYoutubeID(query, resolver)
	default:
		//Search for the query
		if searcher == nil {
			panic("Attempted to search for a query without a searcher")
		}
		return searcher.Search(query, searchSource)
	}
}

func FromSpotifyID(id string, resolver TitleResolver) MusicID {
	mid := MusicID{
		spotifyURI: "spotify:track:" + id[:22],
		SourceName: "spotify",
	}

	return *mid.ResolveTitle(resolver)
}

func FromYoutubeID(id string, resolver TitleResolver) MusicID {
	mid := MusicID{
		YoutubeID:  id[:11],
		SourceName: "youtube",
	}

	return *mid.ResolveTitle(resolver)
}

func (m *MusicID) ResolveTitle(resolver TitleResolver) *MusicID {
	if resolver == nil {
		return m
	}

	if m.Title == "" {
		title, err := resolver.ResolveTitle(m)
		if err != nil {
			log.Printf("Error resolving title for %v: %v\n", m, err)
			return m
		}
		m.Title = title
	}
	return m
}

func (m MusicID) spotify() spotify.URI {
	if m.spotifyURI == "" {
		panic("Attempted to call spotify() on a MusicID without a spotify URI")
	}
	return spotify.URI(m.spotifyURI)
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
	return m.spotifyURI != ""
}

func (m *MusicID) String() string {
	if m.Title != "" {
		return m.Title
	}

	return fmt.Sprintf("Unknown %s track", m.SourceName)
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
