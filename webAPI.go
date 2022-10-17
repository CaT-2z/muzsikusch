package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type HttpAPI struct {
	player *Muzsikusch
}

func NewHttpAPI() *HttpAPI {
	return &HttpAPI{
		player: NewMuzsikusch(),
	}
}

func (api *HttpAPI) startServer() {
	api.registerHandles()
	http.ListenAndServe(":8080", nil)
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
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Couldn't parse form1", http.StatusBadRequest)
		return
	}

	querys, ok := r.Form["query"]
	if !ok {
		http.Error(w, "Couldn't parse form2", http.StatusBadRequest)
		return
	}

	query := querys[0]

	/*source, err := r.Form["source"]
	if !err {
		http.Error(w, "Couldn't parse form3", http.StatusBadRequest)
		return
	}*/

	source := "spotify"

	musicid := FromUser(query, api.player, source, api.player)

	err := api.player.Enqueue(musicid)
	if err != nil {
		log.Printf("Failed to enqueue: %v\n", err)
		http.Error(w, "Couldn't enqueue", http.StatusInternalServerError)
	} else {
		//Maybe send something back?
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
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Couldn't parse form1", http.StatusBadRequest)
		return
	}

	volumes, ok := r.Form["volume"]
	if !ok {
		http.Error(w, "Couldn't parse form2", http.StatusBadRequest)
		return
	}

	volume, err := strconv.Atoi(volumes[0])
	if err != nil {
		http.Error(w, "Couldn't parse form3", http.StatusBadRequest)
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
		//Maybe send something back?
		w.WriteHeader(http.StatusCreated)
	}
}

func (api *HttpAPI) registerHandles() {
	queueEndpoint := EmptyEndpoint().WithGet(api.getQueue).WithPost(api.addToQueue)
	http.Handle("/api/queue", queueEndpoint)

	http.Handle("/api/resume", SimpleEndpoint(api.player.Resume))
	http.Handle("/api/pause", SimpleEndpoint(api.player.Pause))
	http.Handle("/api/skip", SimpleEndpoint(api.player.Skip))
	http.Handle("/api/mute", SimpleEndpoint(api.player.Mute))
	http.Handle("/api/stop", SimpleEndpoint(api.player.Stop))
	//TODO: Seek
	http.Handle("/api/volume", GetEndpoint(api.getVolume).WithPost(api.setVolume))
	http.Handle("/api/", http.NotFoundHandler())

	http.Handle("/", http.FileServer(http.Dir("html")))

}
