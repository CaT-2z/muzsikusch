package endpoints

import (
	"encoding/json"
	"log"
	"muzsikusch/player"
	"muzsikusch/queue"
	"muzsikusch/websocket"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	mplayer   *player.Muzsikusch
	wsmanager = websocket.NewManager()
)

// Takes in a subrouter where it will start matching
func NewV2APIRouter(r *mux.Router, htapi *HttpAPI) *mux.Router {
	mplayer = htapi.Player
	r.HandleFunc("/search", searchHandler).Methods("GET")
	r.HandleFunc("/ws", wsmanager.ServeWS).Schemes("ws", "wss")
	return r
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
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

	MusicIDResults := mplayer.FromUser(query)

	// If only 1 result found, it will enque that
	//TODO: feels like a bad idea
	if len(MusicIDResults) == 0 {
		http.Error(w, "No results", http.StatusInternalServerError)
	}

	if len(MusicIDResults) == 1 {
		err = mplayer.Enqueue(MusicIDResults[0])
		if err != nil {
			log.Printf("Failed to enqueue: %v\n", err)
			http.Error(w, "Couldn't enqueue", http.StatusInternalServerError)
		} else {

		}
	}

	data, err := json.Marshal(MusicIDResults)
	if err != nil {
		http.Error(w, "Couldn't marshal search results", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func appendHandler(w http.ResponseWriter, r *http.Request) {
	var req queue.MusicID
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}
	err = mplayer.Enqueue(req)
	if err != nil {
		log.Printf("Failed to enqueue: %v\n", err)
		http.Error(w, "Couldn't enqueue", http.StatusInternalServerError)
	} else {
		// Maybe send something back?
		w.WriteHeader(http.StatusCreated)
		//TODO: ws write everyone
	}
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	var req queue.MusicID
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}
	err = mplayer.Push(req)
	if err != nil {
		log.Printf("Failed to enqueue: %v\n", err)
		http.Error(w, "Couldn't enqueue", http.StatusInternalServerError)
	} else {
		// Maybe send something back?
		w.WriteHeader(http.StatusCreated)
		//TODO: ws write everyone
	}
}

func forceHandler(w http.ResponseWriter, r *http.Request)  {}
func removeHandler(w http.ResponseWriter, r *http.Request) {}
func skipHandler(w http.ResponseWriter, r *http.Request)   {}

func pauseHandler(w http.ResponseWriter, r *http.Request)   {}
func unpauseHandler(w http.ResponseWriter, r *http.Request) {}

func wsHandler(w http.ResponseWriter, r *http.Request) {}
