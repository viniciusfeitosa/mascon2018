package main

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	createUsersQueue = "CREATE_USER"
	updateUsersQueue = "UPDATE_USER"
	deleteUsersQueue = "DELETE_USER"
	portAddr         = ":50051"
)

func main() {
	connectionString := os.Getenv("DATABASE_URL")

	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	s := Service{}
	s.Initialize(db)
	s.initializeRoutes()
	s.Run(":3000")
}
