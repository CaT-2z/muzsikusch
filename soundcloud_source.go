package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type SoundcloudSource struct {
	MpvSource
	oauth string
}

func NewSoundcloudSource() (src *SoundcloudSource, name string, err error) {

	name = "soundcloud"

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("NewSoundcloudSource: Unable to start soundcloud service: %s", e)
		}
	}()

	mpv, err := NewMpvSource("/tmp/mpvsocket", context.Background())
	if err != nil {
		panic(err)
	}

	tokenFile, err := os.Open(os.Getenv("SOUNDCLOUD_TOKEN_PATH"))
	if err != nil {
		panic(err)
	}

	tokenBytes, err := io.ReadAll(tokenFile)
	if err != nil {
		panic(err)
	}

	var token struct {
		Token string `json:"token"`
	}

	json.Unmarshal(tokenBytes, &token)

	src = &SoundcloudSource{
		MpvSource: *mpv,
		oauth:     token.Token,
	}

	err = src.CheckOAuth()
	if err != nil {
		panic(err)
	}

	go src.waitForEnd(context.Background())

	return
}

func (c *SoundcloudSource) CheckOAuth() (err error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api-widget.soundcloud.com/resolve?url=https://api.soundcloud.com/tracks/70796888&format=json", nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", c.oauth)

	res, err := client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("Couldn't verify OAuth, response code: %d", res.StatusCode)
	}

	return

}

func (c *SoundcloudSource) Play(m MusicID) error {

	url, err := c.GetStreamURL(m)
	if err != nil {
		return fmt.Errorf("Soundcloud couldnt get StreamURL: %s", err)
	}

	_, err = c.instance.Call("loadfile", url)
	if err == nil {
		c.isActive = true
	}
	return err

}

func (c *SoundcloudSource) GetStreamURL(m MusicID) (url string, err error) {

	if m.SourceName != "soundcloud" {
		panic("Tried to get streamURL of non soundcloud track")
	}

	trackInfo, err := c.GetTrackInfo("https://api.soundcloud.com/tracks/" + m.TrackID)
	if err != nil {
		return
	}

	client := &http.Client{}

	request, err := http.NewRequest("GET", trackInfo.Media.Transcodings[1].URL+"?track_authorization="+trackInfo.TrackAuthorization, nil)
	if err != nil {
		return
	}
	request.Header.Set("Authorization", c.oauth)

	res, err := client.Do(request)
	if err != nil {
		return
	}

	all, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	var wrapper struct {
		Url string `json:"url"`
	}
	err = json.Unmarshal(all, &wrapper)
	if err != nil {
		return
	}
	url = wrapper.Url
	return
}

func (c *SoundcloudSource) GetTrackInfo(url string) (info SoundcloudTrackInfo, err error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", "https://api-widget.soundcloud.com/resolve?url="+url+"&format=json", nil)
	if err != nil {
		return
	}
	request.Header.Set("Authorization", c.oauth)

	res, err := client.Do(request)
	if err != nil {
		return
	}

	all, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(all, &info)
	if err != nil {
		return
	}

	res.Body.Close()
	return
}

func (c *SoundcloudSource) Search(query string) []MusicID {
	client := &http.Client{}

	request, err := http.NewRequest("GET", "https://api-v2.soundcloud.com/search?q="+url.QueryEscape(query)+"&limit=5", nil)
	if err != nil {
		return []MusicID{}
	}
	request.Header.Set("Authorization", c.oauth)

	res, err := client.Do(request)
	if err != nil {
		return []MusicID{}
	}

	all, err := io.ReadAll(res.Body)
	if err != nil {
		return []MusicID{}
	}

	//A response would have more fields, but this is all we need
	var results struct {
		Collection []SoundcloudTrackInfo `json:"collection"`
	}

	err = json.Unmarshal(all, &results)
	if err != nil {
		return []MusicID{}
	}

	ret := make([]MusicID, 0)
	for _, song := range results.Collection {
		if song.Kind == "track" {
			ret = append(ret, MusicID{
				TrackID:    song.Urn[len("soundcloud:tracks:"):],
				SourceName: "soundcloud",
				Title:      song.Title,
			})
		}
	}

	res.Body.Close()

	return ret
}

func (c *SoundcloudSource) BelongsToThis(query string) (bool, MusicID) {
	switch {
	case strings.HasPrefix(query, "https://soundcloud.com/"):
		info, err := c.GetTrackInfo(query)
		if err != nil {
			return false, MusicID{}
		}
		return true, MusicID{
			TrackID:    info.Urn[len("soundcloud:tracks:"):],
			SourceName: "soundcloud",
			Title:      info.Title,
		}
	case isSoundcloudID(query):
		info, err := c.GetTrackInfo("https://api.soundcloud.com/tracks/" + query)
		if err != nil {
			return false, MusicID{}
		}
		return true, MusicID{
			TrackID:    info.Urn[len("soundcloud:tracks:"):],
			SourceName: "soundcloud",
			Title:      info.Title,
		}
	default:
		return false, MusicID{}

	}

}

func isSoundcloudID(query string) bool {
	if len(query) != 8 {
		return false
	}
	for _, c := range query {
		if !strings.ContainsRune("0123456789", c) {
			return false
		}
	}
	return true
}

// Completely deprecated, should remove
func (c *SoundcloudSource) ResolveTitle(*MusicID) (string, error) {
	panic("This is not supposed to happen!")
}
