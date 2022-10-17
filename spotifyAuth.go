package main

import (
	"context"
	"log"
	"net/http"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func waitForAuth(state string, auth *spotifyauth.Authenticator) *spotify.Client {
	ch := make(chan *spotify.Client)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(r.Context(), state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}
		// use the token to get an authenticated client
		client := spotify.New(auth.Client(r.Context(), tok))
		http.Redirect(w, r, "/index.html", http.StatusFound)
		ch <- client
	})

	srv := &http.Server{Addr: ":8080"}
	go srv.ListenAndServe()

	client := <-ch
	srv.Shutdown(context.Background())

	return client
}
