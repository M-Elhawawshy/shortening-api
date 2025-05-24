package main

import (
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	standard := alice.New(app.recoverPanic, app.logRequest)

	mux.HandleFunc("/api/auth/signup", app.signUpHandler)
	mux.HandleFunc("/api/auth/login", app.loginHandler)
	mux.HandleFunc("/api/auth/logout", app.logoutHandler)
	mux.HandleFunc("/api/auth/refresh", app.refreshHandler)
	mux.HandleFunc("/api/auth/public.pem", app.pubKeyHandler)

	return standard.Then(mux)
}
