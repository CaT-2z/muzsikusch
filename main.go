package main

import (
	"muzsikusch/endpoints"
	"muzsikusch/middleware"
	"muzsikusch/source"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	// os.Setenv("SOUNDCLOUD_TOKEN_PATH", "./soundcloud_token.json")
	// os.Setenv("YOUTUBE_TOKEN_PATH", "./youtube_token.json")
	// os.Setenv("AUTHSCH_TOKEN_PATH", "./autsch_token.json")

	middleware.SetupAuthSCH()
	middleware.SessionsInit()
	api := endpoints.NewHttpAPI()
	api.Player.SetupSource(source.NewSpotifyFromToken(os.Getenv("SPOTIFY_TOKEN_PATH")))
	api.Player.SetupSource(source.NewYoutubeSource())
	api.Player.SetupSource(source.NewSoundcloudSource())

	api.StartServer()
}
