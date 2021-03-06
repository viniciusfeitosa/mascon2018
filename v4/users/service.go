package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	pb "github.com/viniciusfeitosa/mascon2018/v4/users/preferences"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Service is the struct with app configuration values
type Service struct {
	DB     *sqlx.DB
	Cache  Cache
	Router *mux.Router
}

// Initialize create the DB connection and prepare all the routes
func (s *Service) Initialize(cache Cache, db *sqlx.DB) {
	s.DB = db
	s.Cache = cache
	s.Router = mux.NewRouter()
}

func (s *Service) initializeRoutes() {
	s.Router.HandleFunc("/all", s.getUsers).Methods("GET")
	s.Router.HandleFunc("/", s.createUser).Methods("POST")
	s.Router.HandleFunc("/{id:[0-9]+}", s.getUser).Methods("GET")
	s.Router.HandleFunc("/{id:[0-9]+}/preferences", s.getUserWithPreferences).Methods("GET")
	s.Router.HandleFunc("/{id:[0-9]+}", s.updateUser).Methods("PUT")
	s.Router.HandleFunc("/{id:[0-9]+}", s.deleteUser).Methods("DELETE")
	s.Router.HandleFunc("/healthcheck", s.healthcheck).Methods("GET")
}

// Run initialize the server
func (s *Service) Run(sddr string) {
	n := negroni.Classic()
	n.UseHandler(s.Router)
	log.Fatal(http.ListenAndServe(sddr, n))
}

func (s *Service) healthcheck(w http.ResponseWriter, r *http.Request) {
	var err error
	c := s.Cache.Pool.Get()
	defer c.Close()

	// Check Cache
	_, err = c.Do("PING")

	// Check DB
	err = s.DB.Ping()

	if err != nil {
		http.Error(w, "CRITICAL", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("OK"))
	return
}

func (s *Service) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
	}

	if value, err := s.getUserFromCache(id); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(value))
		return
	}

	user, err := s.getUserFromDB(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			http.Error(w, "User not found", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (s *Service) getUserWithPreferences(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
	}

	user, err := s.getUserFromDB(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			http.Error(w, "User not found", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	conn, err := grpc.Dial(os.Getenv("PREFERENCE_ADDRESS"), grpc.WithInsecure())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer conn.Close()
	c := pb.NewGetPreferenceDataClient(conn)
	response, _ := c.GetPreference(context.Background(), &pb.PreferenceDataRequest{Id: int32(user.ID)})

	data := struct {
		User        User        `json:"user"`
		Preferences interface{} `json:"preferences"`
	}{
		User:        user,
		Preferences: response.Preferences,
	}
	respondWithJSON(w, http.StatusOK, data)
}

func (s *Service) getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	users, err := list(s.DB, start, count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (s *Service) createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
	}
	defer r.Body.Close()

	s.DB.Get(&user.ID, "SELECT nextval('users_id_seq')")
	if err := user.create(s.DB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	JSONByte, _ := json.Marshal(user)
	if err := s.Cache.setValue(user.ID, string(JSONByte)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (s *Service) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
	}

	var user User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Invalid resquest payload", http.StatusBadRequest)
	}
	defer r.Body.Close()
	user.ID = id

	if err := user.update(s.DB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (s *Service) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
	}

	user := User{ID: id}
	if err := user.delete(s.DB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (s *Service) getUserFromCache(id int) (string, error) {
	if value, err := s.Cache.getValue(id); err == nil && len(value) != 0 {
		return value, err
	}
	return "", errors.New("Not Found")
}

func (s *Service) getUserFromDB(id int) (User, error) {
	user := User{ID: id}
	if err := user.get(s.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			return user, err
		default:
			return user, err
		}
	}
	return user, nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
