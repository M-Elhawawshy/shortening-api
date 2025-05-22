package main

import (
	"fmt"
	"log"
	"shortening-api/internal/helpers"
)

func main() {
	_, err := helpers.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	port, err := helpers.GetPort("AUTH_PORT")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(port)

}
