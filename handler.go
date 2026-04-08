package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// function to orchastrate the url shortening
func (app *AppEnv) ShortenURL(w http.ResponseWriter, r *http.Request) {
	long := r.FormValue("url")
	alias := r.FormValue("alias")
	expiry := r.FormValue("expiry")

	var expiresAt *time.Time

	//checks for empty strings
	if long == "" {
		http.Error(w, "Bad Data", http.StatusBadRequest)
		return
	}

	//checks if url format is correct using net/url library
	u, err := url.Parse(long)
	if err != nil || u.Scheme == "" || u.Host == "" {
		http.Error(w, "Enter correct URL", http.StatusBadRequest)
		return
	}

	//Expiry date format Validation
	if expiry != "" {
		t, err := time.Parse("2006-01-02", expiry)
		if err != nil {
			http.Error(w, "Invalid Date Format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}
	//checking if Time is in future
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		http.Error(w, "Expiry date must be in future", http.StatusBadRequest)
		return
	}

	//check if alias is given
	if alias == "" {
		//check if url already exists in the db
		var shortExist string
		err = app.DB.QueryRow("SELECT short_url FROM urls WHERE original_url =$1", long).Scan(&shortExist)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "something went Wrong", http.StatusInternalServerError)
			return
		}
		if shortExist != "" {
			fmt.Fprintf(w, "Your Short URL: http://localhost:8080/%s", shortExist)
		} else {
			smallByte := hashing(long)
			short := encoder(smallByte)
			_, err = app.DB.Exec("INSERT INTO urls (short_url,original_url,expires_at) VALUES ($1,$2,$3)", short, long, expiresAt)
			if err != nil {
				http.Error(w, "something went Wrong", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "Your Short URL: http://localhost:8080/%s", short)
		}
	} else {
		//check if alias is already taken
		var aliasExist string
		err = app.DB.QueryRow("SELECT short_url FROM urls WHERE short_url =$1", alias).Scan(&aliasExist)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "something went Wrong", http.StatusInternalServerError)
			return
		}
		if aliasExist != "" {
			fmt.Fprintf(w, "Custom URL already Taken!")
			return
		}
		_, err := app.DB.Exec("INSERT INTO urls (short_url,original_url,expires_at) VALUES ($1,$2,$3)", alias, long, expiresAt)
		if err != nil {
			http.Error(w, "something went Wrong", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Your Short URL: http://localhost:8080/%s", alias)
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
