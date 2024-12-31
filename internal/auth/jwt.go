package auth

import (
	"fmt"
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
	if a.secret == "" {
		return "", fmt.Errorf("secret key not valid")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func (a *JWTAUthenticator) ValidateToken(token string) (*jwt.Token, error) {

	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v ", t.Header["alg"])
		}

		return []byte(a.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.aud),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
