package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const port string = ":3000"

// Topic is a structure that represents a topic choiced by an user
type Topic struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Preferences is a structure responsible to create a relationship between an User and a group of topics
type Preferences struct {
	UserID         string  `json:"user_id"`
	FavoriteTopics []Topic `json:"favorite_topics"`
}

func (p Preferences) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topics := []Topic{
		Topic{Title: "I love Go!!!", Description: "Go is a programing language very efficient and powerful"},
		Topic{Title: "I love Microservices!!!", Description: "Work using microservices is a good choice for big applications"},
	}
	pref := Preferences{UserID: vars["user_id"], FavoriteTopics: topics}
	response, err := json.Marshal(pref)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func main() {
	m := mux.NewRouter()
	m.Handle("/user/{user_id:[0-9]+}", Preferences{}).Methods("GET")
	log.Println("Server running on", port)
	http.ListenAndServe(port, m)
}
