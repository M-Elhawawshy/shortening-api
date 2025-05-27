package main

import (
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	standard := alice.New(app.recoverPanic, app.logRequest)

	mux.HandleFunc("GET /", app.redirectHandler)

	return standard.Then(mux)
}
