package authentication

import (
	"avito_shop/internal/models"
	"errors"
	"time"
)

type TokenClaims struct {
	Employee struct {
		Username string `json:"username"`
	} `json:"user"`
	IAT int64 `json:"iat"`
	EXP int64 `json:"exp"`
}

type RegisterResponse struct {
	AccessToken string `json:"token"`
}

func GenerateTokenClaims(employee *models.Employee, now time.Time) *TokenClaims {
	username := employee.Username

	newTokenClaims := &TokenClaims{
		Employee: struct {
			Username string `json:"username"`
		}{
			Username: username,
		},
		IAT: now.Unix(),
		EXP: now.Add(time.Hour * 12).Unix(),
	}

	return newTokenClaims
}

func (c *TokenClaims) Valid() error {
	currentTime := time.Now().Unix()

	if c.EXP < currentTime {
		return errors.New("токен истек")
	}

	return nil
}

var SecretKey = []byte("mykey")
