package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const port = ":8080"

// MessageToUser is to return some value strutured
type MessageToUser struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World")
	})

	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		response, err := json.Marshal(MessageToUser{User: name, Message: "Hello World"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})

	log.Println("Run Server on", port)
	http.ListenAndServe(port, nil)
}
