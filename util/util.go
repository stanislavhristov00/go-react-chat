package util

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func LoadEnvFile() error {
	// Check if it the .env file has not already been loaded
	if _, exists := os.LookupEnv("DB_NAME"); !exists {
		currentDir, err := os.Getwd()

		if err != nil {
			log.Fatalf("Couldn't get current working directory: %v", err)
		}

		for {
			// Check if the go.mod file exists in the current directory
			if _, err := os.Stat(filepath.Join(currentDir, ".env")); err == nil {
				break
			}

			// Move up one directory level
			parent := filepath.Dir(currentDir)
			if parent == currentDir {
				// If the parent directory is the same as the current, we've reached the root
				// but still the .env file is nowhere to be found
				return fmt.Errorf("Reached the root directory, but couldn't find .env file")
			}

			currentDir = parent
		}

		envPath := filepath.Join(currentDir, ".env")

		err = godotenv.Load(envPath)

		if err != nil {
			log.Fatalf("Couldn't load .env file: %v", err)
		}

		return err
	}

	return nil
}

func HashPassword(password string) (string, error) {
	// Generate a salted hash using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Convert the hash to a string and return it
	return string(hash), nil
}

// CheckPasswordHash compares a bcrypt hashed password with its possible plaintext equivalent.
func CheckPasswordHash(password, hash string) bool {
	// Compare the hashed password with the plain-text password
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Middleware for a request handler to implement rate limiting
func RateLimitMiddleware(callback func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := CheckForRateLimit(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		callback(w, r)
	})
}

// Function to check the auth header and return the JWT token, if there is one,
// otherwise return an empty string
func GetAuthHeader(request *http.Request) (string, error) {
	// Get the Authorization header value
	authHeader := request.Header.Get("Authorization")

	if authHeader == "" {
		return "", fmt.Errorf("no JWT token supplied in the Authorization header")
	}

	// Parse the contents, since it will contain the "Bearer" prefix
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("malformed Authorization header")
	}

	return parts[1], nil
}
