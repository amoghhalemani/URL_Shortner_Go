package main

import (
	"net/http"
)

func (app *AppEnv) ShortenURL(w http.ResponseWriter, r *http.Request) {
	long := r.FormValue("url")
	smallByte := hashing(long)
	short := encoder(smallByte)
	_, err := app.DB.Exec("INSERT INTO urls (short,long) VALUES (?,?)", short, long)
	if err != nil {
		http.Error(w, "somenthing went Wrong", http.StatusInternalServerError)
		return
	}
}
