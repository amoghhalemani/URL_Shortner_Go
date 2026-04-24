package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Helper for sending JSON errors
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// This creates {"message": "Your error text"}
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}

// Helper for sending JSON success responses
func sendJSONSuccess(w http.ResponseWriter, shortUrl string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// This creates {"shortUrl": "http://localhost:8080/alias"}
	json.NewEncoder(w).Encode(map[string]string{
		"shortUrl": shortUrl,
	})
}

// function to orchastrate the url shortening
func (app *AppEnv) ShortenURL(w http.ResponseWriter, r *http.Request) {
	long := r.FormValue("url")
	alias := r.FormValue("alias")
	expiry := r.FormValue("expiry")

	var expiresAt *time.Time

	//checks for empty strings
	if long == "" {
		sendJSONError(w, "Bad Data", http.StatusBadRequest)
		return
	}

	//checks if url format is correct using net/url library
	u, err := url.Parse(long)
	if err != nil || u.Scheme == "" || u.Host == "" {
		sendJSONError(w, "Enter correct URL", http.StatusBadRequest)
		return
	}

	//Expiry date format Validation
	if expiry != "" {
		t, err := time.Parse("2006-01-02", expiry)
		if err != nil {
			sendJSONError(w, "Invalid Date Format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}
	//checking if Time is in future
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		sendJSONError(w, "Expiry date must be in future", http.StatusBadRequest)
		return
	}

	//check if alias is given
	if alias == "" {
		//check if url already exists in the db
		var shortExist string
		err = app.DB.QueryRow("SELECT short_url FROM urls WHERE original_url =$1", long).Scan(&shortExist)
		if err != nil && err != sql.ErrNoRows {
			sendJSONError(w, "something went Wrong", http.StatusInternalServerError)
			return
		}
		if shortExist != "" {
			sendJSONSuccess(w, "http://localhost:8080/"+shortExist)
		} else {
			smallByte := hashing(long)
			short := encoder(smallByte)
			_, err = app.DB.Exec("INSERT INTO urls (short_url,original_url,expires_at) VALUES ($1,$2,$3)", short, long, expiresAt)
			if err != nil {
				sendJSONError(w, "something went Wrong", http.StatusInternalServerError)
				return
			}
			sendJSONSuccess(w, "http://localhost:8080/"+short)
		}
	} else {
		//check if alias is already taken
		var aliasExist string
		err = app.DB.QueryRow("SELECT short_url FROM urls WHERE short_url =$1", alias).Scan(&aliasExist)
		if err != nil && err != sql.ErrNoRows {
			sendJSONError(w, "something went Wrong", http.StatusInternalServerError)
			return
		}
		if aliasExist != "" {
			sendJSONError(w, "Custom URL already Taken!", http.StatusConflict)
			return
		}
		_, err := app.DB.Exec("INSERT INTO urls (short_url,original_url,expires_at) VALUES ($1,$2,$3)", alias, long, expiresAt)
		if err != nil {
			sendJSONError(w, "something went Wrong", http.StatusInternalServerError)
			return
		}
		sendJSONSuccess(w, "http://localhost:8080/"+alias)
	}
}

// function to redirect to the original url
func (app *AppEnv) redirect(w http.ResponseWriter, r *http.Request) {
	var long string
	var expiresAt sql.NullTime
	short := strings.TrimPrefix(r.URL.Path, "/")
	//checking for Redis hits
	cntxt := r.Context()
	val, err := app.Redis.Get(cntxt, short).Result()
	if err == nil {
		app.Redis.Expire(cntxt, short, 24*time.Hour)
		http.Redirect(w, r, val, http.StatusMovedPermanently)
		return
	}

	//different syntax for Querying in Postgres
	err = app.DB.QueryRow("SELECT original_url,expires_at FROM urls WHERE short_url = $1", short).Scan(&long, &expiresAt)

	if err != nil {
		http.Error(w, "Please Provide the correct url", http.StatusNotFound)
		return
	}

	//checking if link is expired
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		http.Error(w, "Link has Expired", http.StatusGone)
		return
	}

	// Only add to Redis Cache if expiry is not set
	if !expiresAt.Valid {
		app.Redis.Set(cntxt, short, long, 24*time.Hour)
	}
	http.Redirect(w, r, long, http.StatusMovedPermanently)

}
