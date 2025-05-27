package main

import (
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"os"
	"time"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

func init() {
	signBytes, err := os.ReadFile("./config/keys/private_key.pem")
	if err != nil {
		log.Fatal(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}

	verifyBytes, err := os.ReadFile("./config/keys/public_key.pem")
	if err != nil {
		log.Fatal(err)
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func createToken(userID string, expirationDate time.Time) (string, error) {
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims = &jwt.RegisteredClaims{
		Issuer:    "shortening-api auth service",
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expirationDate),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.NewString(),
	}

	return token.SignedString(signKey)
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
}
