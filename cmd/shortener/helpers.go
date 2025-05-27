package main

import "net/http"

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, r *http.Request, err error, status int) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(status), status)
}
