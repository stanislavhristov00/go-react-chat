package auth

import (
	"chat-module/util"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func loadJwtKey() []byte {
	err := util.LoadEnvFile()

	if err != nil {
		return []byte{}
	} else {
		return []byte(os.Getenv("jwtKey"))
	}
}

var jwtKey = loadJwtKey()

type Claims struct {
	username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJWTToken(username string) (string, error) {
	// Set the expiration time for the token
	expirationTime := time.Now().Add(time.Hour * 5)

	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "react-go-chat-app",
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString string) (*Claims, error) {
	// Parse the token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, fmt.Errorf("invalid token signature")
		}
		return nil, fmt.Errorf("error parsing token")
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
