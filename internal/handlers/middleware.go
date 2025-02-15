package handlers

import (
	"avito_shop/internal/authentication"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type contextUsername string

const usernameKey contextUsername = "username"

func (h *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Извлечение токена из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
			return
		}

		// Проверка формата токена
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // Если токен не был найден
			http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
			return
		}

		// Далее идет проверка токена...
		claims := &authentication.TokenClaims{}
		jwtToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неверный метод подписи")
			}
			return authentication.SecretKey, nil
		})

		if err != nil || !jwtToken.Valid {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), usernameKey, claims.Employee.Username)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// func GetUsername(ctx context.Context) (string, bool) {
// 	return "nil", true
// 	//r.Context().Value(usernameKey).(string)
// }
