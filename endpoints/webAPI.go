package endpoints

import (
	"encoding/json"
	"log"
	"muzsikusch/middleware"
	"muzsikusch/player"
	entry "muzsikusch/queue/entry"
	"net/http"

	"github.com/gorilla/mux"
)

type HttpAPI struct {
	Player *player.Muzsikusch
}

func NewHttpAPI() *HttpAPI {
	return &HttpAPI{
		Player: player.NewMuzsikusch(),
	}
}

func (api *HttpAPI) startSecureServer(chainPath, privkeyPath string) {
	api.registerHandles()
	err := http.ListenAndServeTLS(":443", chainPath, privkeyPath, nil)
	if err != nil {
		log.Printf("Failed to serve and listen: %v\n", err)
	}
}

func (api *HttpAPI) StartServer() {
	api.registerHandles()
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Printf("Failed to serve and listen: %v\n", err)
	}
}

func (api *HttpAPI) getQueue(w http.ResponseWriter, r *http.Request) {
	queue := api.Player.GetQueue()

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

	MusicIDResults := api.Player.FromUser(query)

	_, err = json.Marshal(MusicIDResults)
	if err != nil {
		http.Error(w, "Couldn't marshal search results", http.StatusInternalServerError)
	}

	//TODO: this is still bad but its now localised here. Make a better search
	var MusicID entry.MusicID
	if len(MusicIDResults) > 0 {
		MusicID = MusicIDResults[0]
	}
	if len(MusicIDResults) > 1 {
		for _, song := range MusicIDResults {
			if song.SourceName == "spotify" {
				MusicID = song
				break
			}
		}
	}

	err = api.Player.Enqueue(MusicID)
	if err != nil {
		log.Printf("Failed to enqueue: %v\n", err)
		http.Error(w, "Couldn't enqueue", http.StatusInternalServerError)
	} else {
		// Maybe send something back?
		w.WriteHeader(http.StatusCreated)
	}
}

func (api *HttpAPI) getVolume(w http.ResponseWriter, r *http.Request) {
	volume, err := api.Player.GetVolume()
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

	err = api.Player.SetVolume(volume)
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

	http.Handle("/callback", middleware.AuthHandler)

	queueEndpoint := EmptyEndpoint().WithGet(api.getQueue).WithPost(api.addToQueue)
	http.Handle("/api/queue", middleware.AuthRequest(queueEndpoint))

	http.Handle("/api/resume", middleware.AuthRequest(SimpleEndpoint(api.Player.Resume)))
	http.Handle("/api/pause", middleware.AuthRequest(SimpleEndpoint(api.Player.Pause)))
	http.Handle("/api/skip", middleware.AuthRequest(SimpleEndpoint(api.Player.Skip)))
	http.Handle("/api/mute", middleware.AuthRequest(SimpleEndpoint(api.Player.Mute)))
	http.Handle("/api/stop", middleware.AuthRequest(SimpleEndpoint(api.Player.Stop)))
	/* TODO:
	- Seek
	- Search
	- Delete
	*/
	http.Handle("/api/volume", middleware.AuthRequest(GetEndpoint(api.getVolume).WithPost(api.setVolume)))
	http.Handle("/api/", http.NotFoundHandler())

	r := mux.NewRouter()

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("pong")) })
	NewV2APIRouter(r.PathPrefix("/v2/api").Subrouter(), api)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("html")))

	http.Handle("/", middleware.AuthRequest(r))
	//http.Handle("/", http.FileServer(http.Dir("html")))
}
