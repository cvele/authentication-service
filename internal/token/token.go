package token

import (
	"errors"
	"time"

	"github.com/cvele/authentication-service/internal/config"
	"github.com/dgrijalva/jwt-go/v4"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func New(userID string, cfg *config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(cfg.TokenTTL)),
		},
	})

	return token.SignedString([]byte(cfg.JWTSecretKey))
}

func Validate(tokenString string, cfg *config.Config) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(cfg.JWTSecretKey), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	if err := claims.Valid(&jwt.ValidationHelper{}); err != nil {
		return "", err
	}

	return claims.UserID, nil
}
