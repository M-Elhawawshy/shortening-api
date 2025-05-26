package main

import (
	"fmt"
	"github.com/justinas/alice"
	"log"
	"log/slog"
	"net/http"
	"os"
	"shortening-api/internal/database"
	"shortening-api/internal/helpers"
	"strings"
)

type application struct {
	logger  *slog.Logger
	queries *database.Queries
}

func main() {
	db, err := helpers.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	port, err := helpers.GetEnv("REDIRECT_PORT")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	queries := database.New(db)

	app := application{
		logger:  logger,
		queries: queries,
	}
	log.Println("Auth app is listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, app.routes()))
}

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	standard := alice.New(app.recoverPanic, app.logRequest)

	mux.HandleFunc("GET /", app.redirectHandler)

	return standard.Then(mux)
}

func (app *application) redirectHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimPrefix(r.URL.Path, "/")
	fmt.Fprintf(w, "Captured: %s", url)
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, r *http.Request, err error, status int) {
	app.logger.Error(err.Error(), "method: ", r.Method, " uri: ", r.RequestURI)
	http.Error(w, http.StatusText(status), status)
}
