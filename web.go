package main

import (
	"encoding/gob"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	ctph "github.com/joekir/ssdeepviz/src/ctph"
	"log"
	"net/http"
	"os"
)

const (
	FUZZY_HASH_OBJ = "fuzzyHash"
	CTPH_COOKIE    = "ctph"
)

var LISTENING_PORT = os.Getenv("PORT")
var cookieStore *sessions.CookieStore

func init() {
	// Server side storage
	cookieStore = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SESSION_KEY")))
	gob.RegisterName(FUZZY_HASH_OBJ, &ctph.FuzzyHash{})
	gob.Register(&ctph.RollingHash{})
	if len(LISTENING_PORT) < 1 {
		LISTENING_PORT = "8080"
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/NewHash", NewHash).Methods("POST")
	router.HandleFunc("/StepHash", StepHash).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Printf("Listening on %s\n", LISTENING_PORT)
	log.Fatal(http.ListenAndServe(":"+LISTENING_PORT, router))
}

type hashReq struct {
	DataLength int `json:"data_length"`
}

func NewHash(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var h hashReq
	err := decoder.Decode(&h)

	log.Printf("NewHash request received: %#v\n", h)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.DataLength <= 0 {
		http.Error(w, "Invalid 'data_length'", http.StatusUnprocessableEntity)
		return
	}

	// A session is always returned
	session, _ := cookieStore.Get(r, CTPH_COOKIE)

	if !session.IsNew {
		log.Println("Deleting old cookie")
		session.Options.MaxAge = -1
		session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session, err = cookieStore.New(r, CTPH_COOKIE)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fh := ctph.NewFuzzyHash(h.DataLength)
	session.Values[FUZZY_HASH_OBJ] = fh
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fh)
}

type stepReq struct {
	Data byte `json:"byte"`
}

func StepHash(w http.ResponseWriter, r *http.Request) {
	session, err := cookieStore.Get(r, CTPH_COOKIE)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// This should be acting on a server-side struct persisted with gob
	// Hence a session is needed from /NewHash call
	if session.IsNew {
		http.Error(w, "no session detected", http.StatusPreconditionRequired)
		return
	}

	fh_cookie := session.Values[FUZZY_HASH_OBJ]
	fh, ok := fh_cookie.(*ctph.FuzzyHash)
	if !ok {
		http.Error(w, "Unable to retrieve FuzzyHash from session obj",
			http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)

	var s stepReq
	err = decoder.Decode(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.Data == 0x0 {
		// You could argue that 0x0 is a legitimate state, however in ascii it is NUL
		// Hence it's unlikely to be a legit input, however this is a default input if the
		// Client doesn't have a valid one, so we should return
		http.Error(w, "No data provided, no state to update", http.StatusNoContent)
		return
	}

	log.Printf("StepHash request received: %#v\n", s)

	fh.Step(s.Data)
	session.Values[FUZZY_HASH_OBJ] = fh
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fh)
}
