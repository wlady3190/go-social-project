package auth

import (
	"fmt"
	"log"
	"github.com/golang-jwt/jwt/v5"
)

type JWTAUthenticator struct {
	secret string
	aud    string
	iss    string
}

func NewJWTAuthenticator(secret, aud, iss string) *JWTAUthenticator {
	return &JWTAUthenticator{secret, iss, aud}
}

func (a *JWTAUthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	if a.secret ==""{
		log.Print(a.secret)
		return "", fmt.Errorf("llave secreta no valida")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil

}


func (a *JWTAUthenticator) ValidateToken(token string) (*jwt.Token, error){

	return nil, nil
}
