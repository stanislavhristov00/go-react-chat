package main

import (
	"chat-module/db"
	"log"
)

func main() {
	err := db.Init()

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	} else {
		log.Println("We are happy :)")
	}
}
