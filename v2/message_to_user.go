package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// MessageToUser is to return some value strutured
type MessageToUser struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

// ServeHTTP is an interface implementation to communicate something with the internet
func (mtu MessageToUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	response, err := json.Marshal(MessageToUser{User: name, Message: "Hello World"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (mtu MessageToUser) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World")
}
