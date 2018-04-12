package main

import (
	"log"
	"net/http"
)

const port = ":8080"

func main() {

	messageToUser := MessageToUser{}

	http.HandleFunc("/", messageToUser.index)
	http.Handle("/user", messageToUser)

	log.Println("Run Server on", port)
	http.ListenAndServe(port, nil)
}
