package endpoints

import (
	"encoding/json"
	"log"
	"muzsikusch/player"
	entry "muzsikusch/queue/entry"
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
	htapi.Player.SetWSManager(wsmanager)

	r.HandleFunc("/search", searchHandler).Methods("GET")
	r.HandleFunc("/queue", queueHandler).Methods("GET")
	r.HandleFunc("/ws", wsmanager.ServeWS)
	r.HandleFunc("/append", appendHandler).Methods("POST")
	r.HandleFunc("/push", pushHandler).Methods("POST")
	r.HandleFunc("/force", forceHandler).Methods("POST")
	r.HandleFunc("/remove", removeHandler).Methods("DELETE")
	r.HandleFunc("/skip", skipHandler).Methods("DELETE")
	return r
}

func appendHandler(w http.ResponseWriter, r *http.Request) {
	type addRequest struct {
		Query entry.MusicID `json:"musicID"`
	}

	var req addRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}

	err = mplayer.Enqueue(req.Query)
	if err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}

}

func queueHandler(w http.ResponseWriter, r *http.Request) {
	list := mplayer.GetQueue()
	b, e := json.Marshal(list)
	if e != nil {
		http.Error(w, "Couldn't parse queue", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
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

func pushHandler(w http.ResponseWriter, r *http.Request) {
	var req entry.MusicID
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

func forceHandler(w http.ResponseWriter, r *http.Request) {}

func removeHandler(w http.ResponseWriter, r *http.Request) {
	type removeRequest struct {
		UID string `json:"UID"`
	}

	var req removeRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}

	if mplayer.Remove(req.UID) {
		w.WriteHeader(http.StatusAccepted)
	} else {
		http.Error(w, "Couldn't remove", http.StatusBadRequest)
	}
}

func skipHandler(w http.ResponseWriter, r *http.Request) {
	mplayer.Skip()
	w.Write([]byte{})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {}
