package util

import (
	"log"
	"os"
	"path/filepath"

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
