package routes

import (
	"avito_shop/internal/handlers"

	"github.com/gorilla/mux"
)

const (
	Item = "item"
)

func InitRoutes(userHandler *handlers.UserHandler) *mux.Router {
	api := mux.NewRouter()

	api.HandleFunc("/api/auth", userHandler.Auth).Methods("POST")

	// Создание подмаршрута для /api с middleware
	authHandler := api.PathPrefix("/api").Subrouter()
	authHandler.Use(userHandler.AuthMiddleware)

	authHandler.HandleFunc("/buy/{"+Item+"}", userHandler.BuyItem).Methods("GET")
	authHandler.HandleFunc("/sendCoin", userHandler.SendCoin).Methods("POST")
	authHandler.HandleFunc("/info", userHandler.Info).Methods("GET")
	return api
}
