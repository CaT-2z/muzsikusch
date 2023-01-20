package main

import (
	"os"

	"github.com/joho/godotenv"
)

const redirectURI = "http://localhost:8080/callback"

func main() {
	godotenv.Load()
	// os.Setenv("SOUNDCLOUD_TOKEN_PATH", "./soundcloud_token.json")
	// os.Setenv("YOUTUBE_TOKEN_PATH", "./youtube_token.json")
	// os.Setenv("AUTHSCH_TOKEN_PATH", "./autsch_token.json")

	SetupAuthSCH()
	SessionsInit()
	api := NewHttpAPI()
	api.player.SetupSource(NewSpotifyFromToken(os.Getenv("SPOTIFY_TOKEN_PATH")))
	api.player.SetupSource(NewYoutubeSource())
	api.player.SetupSource(NewSoundcloudSource())

	api.startServer()
}
