package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// Middleware to log the requests and response
func logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next(w, r)
	}
}

// middleware for rate limiting
func (app *AppEnv) rateLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		}
		key := fmt.Sprintf("rate:%s", clientIP)
		ctx := r.Context()
		//increasing request count
		counter, err := app.Redis.Incr(ctx, key).Result()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		//checking expiry - set to 1 minute
		if counter == 1 {
			app.Redis.Expire(ctx, key, time.Minute)
		}
		if counter > 10 {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
