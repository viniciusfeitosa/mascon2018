package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	pb "github.com/viniciusfeitosa/mascon2018/v4/preferences/preferences"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
)

const (
	port     = ":3000"
	portAddr = ":50051"
)

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

type preferencesDataHandler struct{}

func runGRPCServer() {
	lis, err := net.Listen("tcp", portAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGetPreferenceDataServer(s, &preferencesDataHandler{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (handler *preferencesDataHandler) GetPreference(ctx context.Context, request *pb.PreferenceDataRequest) (*pb.PreferenceDataResponse, error) {
	data := &pb.PreferenceDataResponse{
		Preferences: []*pb.PreferenceData{
			&pb.PreferenceData{Title: "I love Go!!!", Description: "Go is a programing language very efficient and powerful"},
			&pb.PreferenceData{Title: "I love Microservices!!!", Description: "Work using microservices is a good choice for big applications"},
		},
	}

	return data, nil
}

func main() {
	m := mux.NewRouter()
	m.Handle("/user/{user_id:[0-9]+}", Preferences{}).Methods("GET")
	go runGRPCServer()
	log.Println("Server running on", port)
	http.ListenAndServe(port, m)
}
