package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type HttpAPI struct {
	player *Muzsikusch
}

func NewHttpAPI() *HttpAPI {
	return &HttpAPI{
		player: NewMuzsikusch(),
	}
}

func (api *HttpAPI) startSecureServer(chainPath, privkeyPath string) {
	api.registerHandles()
	err := http.ListenAndServeTLS(":443", chainPath, privkeyPath, nil)
	if err != nil {
		log.Printf("Failed to serve and listen: %v\n", err)
	}
}

func (api *HttpAPI) startServer() {
	api.registerHandles()
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Printf("Failed to serve and listen: %v\n", err)
	}
}

func (api *HttpAPI) getQueue(w http.ResponseWriter, r *http.Request) {
	queue := api.player.GetQueue()

	data, err := json.Marshal(queue)
	if err != nil {
		http.Error(w, "Couldn't marshal queue", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *HttpAPI) addToQueue(w http.ResponseWriter, r *http.Request) {
	type addRequest struct {
		Query string `json:"query"`
	}

	var req addRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}

	query := req.Query
	if query == "" {
		http.Error(w, "No title given", http.StatusBadRequest)
		return
	}

	source := "spotify"

	musicid := FromUser(query, api.player, source, api.player)

	err = api.player.Enqueue(musicid)
	if err != nil {
		log.Printf("Failed to enqueue: %v\n", err)
		http.Error(w, "Couldn't enqueue", http.StatusInternalServerError)
	} else {
		// Maybe send something back?
		w.WriteHeader(http.StatusCreated)
	}
}

func (api *HttpAPI) getVolume(w http.ResponseWriter, r *http.Request) {
	volume, err := api.player.GetVolume()
	if err != nil {
		log.Printf("Failed to get volume: %v\n", err)
		http.Error(w, "Couldn't get volume", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(volume)
	if err != nil {
		http.Error(w, "Couldn't marshal volume", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *HttpAPI) setVolume(w http.ResponseWriter, r *http.Request) {
	var volume int
	err := json.NewDecoder(r.Body).Decode(&volume)
	if err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}

	if volume < 0 || volume > 100 {
		http.Error(w, "Volume out of range", http.StatusBadRequest)
		return
	}

	err = api.player.SetVolume(volume)
	if err != nil {
		log.Printf("Failed to set volume: %v\n", err)
		http.Error(w, "Couldn't set volume", http.StatusInternalServerError)
	} else {
		// Maybe send something back?
		w.WriteHeader(http.StatusCreated)
	}
}

func (api *HttpAPI) registerHandles() {
	passwords, err := os.Open("passwords.json")
	if err != nil {
		panic(err)
	}

	validator, err := NewBasicPasswordValidator(passwords)
	if err != nil {
		panic(err)
	}

	auth := NewBasicAuthenticator("muzsikusch", validator)

	queueEndpoint := EmptyEndpoint().WithGet(api.getQueue).WithPost(api.addToQueue)
	http.Handle("/api/queue", auth.Wrap(queueEndpoint))

	http.Handle("/api/resume", auth.Wrap(SimpleEndpoint(api.player.Resume)))
	http.Handle("/api/pause", auth.Wrap(SimpleEndpoint(api.player.Pause)))
	http.Handle("/api/skip", auth.Wrap(SimpleEndpoint(api.player.Skip)))
	http.Handle("/api/mute", auth.Wrap(SimpleEndpoint(api.player.Mute)))
	http.Handle("/api/stop", auth.Wrap(SimpleEndpoint(api.player.Stop)))
	// TODO: Seek
	http.Handle("/api/volume", auth.Wrap(GetEndpoint(api.getVolume).WithPost(api.setVolume)))
	http.Handle("/api/", http.NotFoundHandler())

	http.Handle("/", http.FileServer(http.Dir("html")))
}
