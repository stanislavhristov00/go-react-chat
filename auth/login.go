package auth

import (
	"chat-module/db"
	"chat-module/models"
	"chat-module/util"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

/*
 *	TODO:
 *  - Make the handlers take a Repo interface for the db operations
 *	- Implemente black-listing of jwt tokens
 *	- add a way to extend the jwt token if a user is nearing his expiration date, but there is still activity going on,
 *    so we don't force him to relogin (figure this out in a safe way, could be dangerous if a malicious user got his hand on
	  the token of another person)
*/

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the body of the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	// Unmarshal the JSON data into a User struct
	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if user.Email == nil || user.Password == nil || user.Username == nil {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	exists, err := db.Client.CheckUserExists(*user.Username, *user.Email)

	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		log.Printf("Failed to query database for user %s: %v", *user.Username, err)
		return
	}

	if exists {
		http.Error(w, "User with this email/username already exists.", http.StatusBadRequest)
		return
	}

	// TODO:
	// We expect the email and password validation to have been done on the front end part
	// Still, do some validation here
	user.ID = primitive.NewObjectID()

	// Hash and salt the password
	hashedPassword, err := util.HashPassword(*user.Password)

	if err == bcrypt.ErrPasswordTooLong {
		http.Error(w, "Password exceeds 72 characters", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Something went wrong with hashing the password", http.StatusInternalServerError)
		log.Printf("Failed to hash password: %v", err)
		return
	}

	user.Password = &hashedPassword

	err = db.Client.AddUser(user)

	if err != nil {
		http.Error(w, "Failed to register user to the database", http.StatusInternalServerError)
		log.Printf("Failed to insert user %s to the database: %v", *user.Username, err)
	}

	// Send a success response
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

/*
 *	Here username can pertain to the actual username of the user or his email,
 * 	so we will check both of these options.
 */
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the body of the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	type LoginUser struct {
		UsernameOrEmail *string `json:"username"`
		Password        *string `json:"password"`
	}

	// Unmarshal the JSON data into a LoginUser struct
	var loginUser LoginUser
	if err := json.Unmarshal(body, &loginUser); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if loginUser.Password == nil || loginUser.UsernameOrEmail == nil {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var user *models.User
	user, err = db.Client.GetUser(*loginUser.UsernameOrEmail)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Such an user doesn't exist", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Failed to find document", http.StatusInternalServerError)
			log.Printf("Failed to find document for user %s: %v", *loginUser.UsernameOrEmail, err)
			return
		}
	}

	// We have such an user, check if the passwords match
	passwordMatches := util.CheckPasswordHash(*loginUser.Password, *user.Password)

	if !passwordMatches {
		http.Error(w, "An user with this user/pass combination doesn't exist", http.StatusBadRequest)
		return
	}

	// We have logged in, generate a JWT token for this user and send it back
	token, err := GenerateJWTToken(*user.Username)

	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		log.Printf("Failed to generate token for user %s: %v", *user.Username, err)
		return
	}

	log.Printf("Generated %s token for %s", token, *user.Username)

	response := map[string]string{
		"token": token,
	}

	// Marshal the map to JSON bytes
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal the response", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Send a success response
	w.WriteHeader(http.StatusCreated)

	// Write the JSON bytes to the response body
	_, err = w.Write(responseBytes)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
