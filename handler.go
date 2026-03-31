package main

import (
	"fmt"
	"net/http"
	"strings"
)

// function to orchastrate the url shortening
func (app *AppEnv) ShortenURL(w http.ResponseWriter, r *http.Request) {
	long := r.FormValue("url")
	smallByte := hashing(long)
	short := encoder(smallByte)
	_, err := app.DB.Exec("INSERT INTO urls (short,long) VALUES (?,?)", short, long)
	if err != nil {
		http.Error(w, "somenthing went Wrong", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Your Short URL: http://localhost:8080/%s", short)
}

// function to redirect to the original url
func (app *AppEnv) redirect(w http.ResponseWriter, r *http.Request) {
	var long string
	short := strings.TrimPrefix(r.URL.Path, "/")
	//different syntax for Querying
	err := app.DB.QueryRow("SELECT long FROM urls WHERE short = ?", short).Scan(&long)
	if err != nil {
		http.Error(w, "Please Provide the correct url", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, long, http.StatusMovedPermanently)
}
