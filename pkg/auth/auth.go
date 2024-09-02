package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	EmailID     string
	ExpiryDate  time.Time
	InitiatedAt time.Time
}

func CreateJWTToken(input string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"emailID": input,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
			"iat":     time.Now().Unix(),
		})
	secretKey := "hiiii" //TODO: read it from env
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWTToken(tokenString string) (*jwt.Token, error) {

	secretKey := "hiiii" //TODO: read it from env
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func ExtractClaims(token *jwt.Token) (*JWTClaims, error) {

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		return &JWTClaims{
			EmailID:     claims["emailID"].(string),
			ExpiryDate:  claims["exp"].(time.Time),
			InitiatedAt: claims["iat"].(time.Time),
		}, nil

	}
	return nil, fmt.Errorf("err")
}
