package main

import (
	"log"
	"net/http"
	"shortening-api/internal/helpers"
)

type application struct {
}

func main() {
	_, err := helpers.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	port, err := helpers.GetEnv("AUTH_PORT")
	if err != nil {
		log.Fatal(err)
	}
	app := application{}
	log.Fatal(http.ListenAndServe(":"+port, app.routes()))
}
