package main

import (
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
	port, err := helpers.GetEnv("AUTH_PORT")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	queries := database.New(db)

	app := application{
		logger:  logger,
		queries: queries,
	}
	app.logger.Info("Auth app is listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, app.routes()))
}
