package main

import (
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	standard := alice.New(app.recoverPanic, app.logRequest)

	mux.HandleFunc("/signup", app.signUpHandler)
	mux.HandleFunc("/login", app.loginHandler)
	mux.HandleFunc("/logout", app.logoutHandler)
	mux.HandleFunc("/refresh", app.refreshHandler)
	mux.HandleFunc("/public.pem", app.pubKeyHandler)

	return standard.Then(mux)
}
