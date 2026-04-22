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

	"github.com/getsentry/sentry-go"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// AppEnv holds the application environment, which includes the database connection.
type AppEnv struct {
	DB    *sql.DB
	Redis *redis.Client
}

func main() {

	//Sentry for Error Monitoring
	godotenv.Load()
	err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	//initializing the database
	dburl := os.Getenv("DB_URL")
	log.Printf("Connecting to DB: %s", dburl)
	db, err := InitDB(dburl)
	if err != nil {
		log.Fatal(err)
	}

	//initializing Redis Client
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(opt)
	err = rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal(err)
	}

	//checking if Table exist and creating if it doesnt
	err = CreateTable(db)
	if err != nil {
		log.Fatal(err)
	}
	//checking if index exist and creating if it doesnt
	err = CreateIndex(db)
	if err != nil {
		log.Fatal(err)
	}

	//creating the app Struct
	app := &AppEnv{DB: db, Redis: rdb}

	//registering the routes
	http.HandleFunc("/shorten", logging(app.rateLimiter(app.ShortenURL)))
	http.HandleFunc("/", logging(app.rateLimiter(app.redirect)))

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

	//gracefull shutdown
	cntxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Printf("Server Shutting Down Gracefully...")
	err = server.Shutdown(cntxt)
	if err != nil {
		log.Fatal(err)
	}
}
