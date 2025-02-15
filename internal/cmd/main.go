package main

import (
	"avito_shop/internal/core"
	"avito_shop/internal/handlers"
	"avito_shop/internal/routes"
	"avito_shop/internal/storage"
	"context"
	"fmt"
	"log"
	"net/http"

	gorillaHandlers "github.com/gorilla/handlers"
)

func main() {
	cfg := storage.PostgresConnConfig{
		DBHost:   "db",
		DBPort:   5432,
		DBName:   "shop",
		Username: "postgres",
		Password: "password",
		Options:  nil,
	}

	// Создание соединения с базой данных
	conn, err := storage.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer conn.Close(context.Background())

	avitoDB := storage.NewAvitoDB(conn)
	authService := core.New(avitoDB)
	userHandler := handlers.NewUserHandler(authService)

	mux := routes.InitRoutes(userHandler)

	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}), // Разрешить все домены, но лучше указать конкретные
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	fmt.Println("Запуск сервера на порту 8080 http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", corsHandler(mux)))
}
