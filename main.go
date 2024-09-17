package main

import (
	"chat-module/auth"
	"chat-module/db"
	"chat-module/util"
	"log"
	"net/http"
	"time"
)

func main() {
	err := db.Init()

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	} else {
		log.Println("We are happy :)")
	}

	http.Handle("/login", util.RateLimitMiddleware(auth.LoginHandler))
	http.Handle("/register", util.RateLimitMiddleware(auth.RegisterHandler))

	go util.RateLimit()

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
