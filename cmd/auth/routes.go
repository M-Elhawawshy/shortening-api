package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/signup", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/api/auth/logout", func(w http.ResponseWriter, r *http.Request) {})

	return mux
}
