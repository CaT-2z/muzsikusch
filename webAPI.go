package main

import (
	"encoding/json"
	"log"
	"net/http"
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

	list := make([]struct{ Title string }, len(queue))
	for i, entry := range queue {
		list[i] = struct{ Title string }{Title: entry.Title}
	}

	data, err := json.Marshal(list)
	if err != nil {
		http.Error(w, "Couldn't marshal queue", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *HttpAPI) search(w http.ResponseWriter, r *http.Request) {
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

	musicidResults := FromUser(query, api.player)

	if len(musicidResults) == 0 {
		http.Error(w, "No results", http.StatusInternalServerError)
	}

	if len(musicidResults) == 1 {
		err = api.player.Enqueue(musicidResults[0])
		if err != nil {
			log.Printf("Failed to enqueue: %v\n", err)
			http.Error(w, "Couldn't enqueue", http.StatusInternalServerError)
		} else {

		}
	}

	data, err := json.Marshal(musicidResults)
	if err != nil {
		http.Error(w, "Couldn't marshal search results", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// This will replace add to queue
func (api *HttpAPI) newQueue(w http.ResponseWriter, r *http.Request) {
	var req MusicID
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}
	err = api.player.Enqueue(req)
	if err != nil {
		log.Printf("Failed to enqueue: %v\n", err)
		http.Error(w, "Couldn't enqueue", http.StatusInternalServerError)
	} else {
		// Maybe send something back?
		w.WriteHeader(http.StatusCreated)
	}
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

	// I moved the part that separates out a single spotify track here, will still need to remove sometime

	musicidResults := FromUser(query, api.player)

	_, err = json.Marshal(musicidResults)
	if err != nil {
		http.Error(w, "Couldn't marshal search results", http.StatusInternalServerError)
	}

	//TODO: this is still bad but its now localised here. Make a better search
	var musicid MusicID
	if len(musicidResults) > 0 {
		musicid = musicidResults[0]
	}
	if len(musicidResults) > 1 {
		for _, song := range musicidResults {
			if song.SourceName == "spotify" {
				musicid = song
				break
			}
		}
	}

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
	//Maybe have a separate search and queue function?

	http.Handle("/callback", AuthHandler)

	queueEndpoint := EmptyEndpoint().WithGet(api.getQueue).WithPost(api.addToQueue)
	http.Handle("/api/queue", AuthRequest(queueEndpoint))

	http.Handle("/api/resume", AuthRequest(SimpleEndpoint(api.player.Resume)))
	http.Handle("/api/pause", AuthRequest(SimpleEndpoint(api.player.Pause)))
	http.Handle("/api/skip", AuthRequest(SimpleEndpoint(api.player.Skip)))
	http.Handle("/api/mute", AuthRequest(SimpleEndpoint(api.player.Mute)))
	http.Handle("/api/stop", AuthRequest(SimpleEndpoint(api.player.Stop)))
	// TODO: Seek
	http.Handle("/api/volume", AuthRequest(GetEndpoint(api.getVolume).WithPost(api.setVolume)))
	http.Handle("/api/", http.NotFoundHandler())

	http.Handle("/", AuthRequest(http.FileServer(http.Dir("html"))))
}
