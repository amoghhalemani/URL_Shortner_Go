package main

// This file initializes the application environment and starts the server.

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

// AppEnv holds the application environment, which includes the database connection.
type AppEnv struct {
	DB *sql.DB
}

func main() {

	//initializing the database
	db, err := InitDB("urls.db")
	if err != nil {
		log.Fatal(err)
	}

	//checking if Table exist and creating if it doesnt
	err = CreateTable(db)
	if err != nil {
		log.Fatal(err)
	}

	//creating the app Struct
	app := &AppEnv{DB: db}

	//registering the routes
	http.HandleFunc("/shorten", logging(app.ShortenURL))
	http.HandleFunc("/", logging(app.redirect))

	//telling Go to listen for requests
	log.Fatal(http.ListenAndServe(":8080", nil))
}
