package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"log/slog"
	"net/http"
	"os"
	"shortening-api/internal/database"
	"shortening-api/internal/helpers"
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
	port, err := helpers.GetEnv("GATEWAY_PORT")
	if err != nil {
		log.Fatal(err)
	}
	authPort, err := helpers.GetEnv("AUTH_PORT")
	if err != nil {
		log.Fatal(err)
	}
	redirectPort, err := helpers.GetEnv("REDIRECT_PORT")
	if err != nil {
		log.Fatal(err)
	}
	shortenerPort, err := helpers.GetEnv("SHORTENER_PORT")
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
	r := chi.NewRouter()
	r.Use(app.logRequest, app.recoverPanic)
	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", app.proxyHandler("http://localhost:"+authPort))
		// use auth middleware
		r.Route("/shorten", func(r chi.Router) {
			r.Use(app.authMiddleware)
			r.Handle("/*", app.proxyHandler("http://localhost:"+shortenerPort))
		})
		r.Route("/redirect", func(r chi.Router) {
			r.Use(app.authMiddleware)
			r.Handle("/*", app.proxyHandler("http://localhost:"+redirectPort))
		})
	})
	log.Println("Auth app is listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
