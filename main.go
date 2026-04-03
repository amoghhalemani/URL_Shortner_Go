package main

// This file initializes the application environment and starts the server.

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	server := &http.Server{
		Addr:         ":8080",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Printf("Server Ready to Listen to Requests...")
	go server.ListenAndServe()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cntxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Printf("Server Shutting Down Gracefully...")
	err = server.Shutdown(cntxt)
	if err != nil {
		log.Fatal(err)
	}
}
