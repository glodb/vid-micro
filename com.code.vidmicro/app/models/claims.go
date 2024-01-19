package models

import "github.com/dgrijalva/jwt-go"

type Claims struct {
	User User
	jwt.StandardClaims
}
