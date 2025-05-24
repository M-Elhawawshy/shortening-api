package main

import "net/http"

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func (app *application) unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
