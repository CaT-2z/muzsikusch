package main

import (
	"muzsikusch/src/endpoints"
	"muzsikusch/src/middleware"
	source2 "muzsikusch/src/source"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	middleware.SetupAuthSCH()
	middleware.SessionsInit()
	api := endpoints.NewHttpAPI()
	api.Player.SetupSource(source2.NewSpotifyFromToken(os.Getenv("SPOTIFY_TOKEN_PATH")))
	api.Player.SetupSource(source2.NewYoutubeSource())
	api.Player.SetupSource(source2.NewSoundcloudSource())

	api.StartServer()
}
