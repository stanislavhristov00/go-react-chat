package test

import (
	"chat-module/auth"
	"chat-module/util"
	"log"
	"net/http"
)

func Test200ResponseHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request to TEST endpoint from user: %s", r.RemoteAddr)

	token, err := util.GetAuthHeader(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	_, err = auth.ValidateJWT(token)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("TEST endpoint OK"))
}
