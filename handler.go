package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// function to orchastrate the url shortening
func (app *AppEnv) ShortenURL(w http.ResponseWriter, r *http.Request) {
	long := r.FormValue("url")

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

	smallByte := hashing(long)
	short := encoder(smallByte)
	_, err = app.DB.Exec("INSERT INTO urls (short_url,original_url) VALUES ($1,$2)", short, long)
	if err != nil {
		http.Error(w, "something went Wrong", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Your Short URL: http://localhost:8080/%s", short)
}

// function to redirect to the original url
func (app *AppEnv) redirect(w http.ResponseWriter, r *http.Request) {
	var long string
	short := strings.TrimPrefix(r.URL.Path, "/")
	//different syntax for Querying
	err := app.DB.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", short).Scan(&long)
	if err != nil {
		http.Error(w, "Please Provide the correct url", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, long, http.StatusMovedPermanently)
}
