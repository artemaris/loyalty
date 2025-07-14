package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var jwtKey = []byte("secret")

func GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
}

func ParseToken(tokenStr string) (int, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	return int(claims["sub"].(float64)), nil
}
