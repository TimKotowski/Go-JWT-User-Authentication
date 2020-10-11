package jwtgenerate

import (
	"log"

	"github.com/dgrijalva/jwt-go"
)

var SECRET_KEY = []byte("gosecretkey")

func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		log.Fatalf(" \n err in jwt generate token %v", err)
		return "", err
	}
	return tokenString, nil
}
