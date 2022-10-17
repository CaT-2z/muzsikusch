package main

import (
	"os"

	"github.com/dexterlb/mpvipc"
)

const redirectURI = "http://localhost:8080/callback"

func main() {
	mpv := mpvipc.NewConnection("/tmp/mpvsocket")
	err := mpv.Open()
	if err != nil {
		panic(err)
	}

	spotSource := NewSpotifyFromToken(os.Getenv("SPOTIFY_TOKEN_PATH"))
	//spotSource := NewSpotifyWithAuth()
	defer spotSource.SaveToken(os.Getenv("SPOTIFY_TOKEN_PATH"))
	youtubeSource := NewYoutubeSource()

	api := NewHttpAPI()
	api.player.RegisterSource(spotSource, "spotify")
	api.player.RegisterSource(youtubeSource, "youtube")
	api.player.RegisterResolver(spotSource, "spotify")
	api.player.RegisterResolver(youtubeSource, "youtube")

	spotSource.Register(api.player.OnPlaybackFinished)
	youtubeSource.Register(api.player.OnPlaybackFinished)

	api.startServer()
}
