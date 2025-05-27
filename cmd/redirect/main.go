package main

import (
	"github.com/redis/go-redis/v9"
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
	cache   *redis.Client
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

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	app := application{
		logger:  logger,
		queries: queries,
		cache:   client,
	}

	log.Println("redirect service is listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, app.routes()))
}
