package buyitem

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"avito_shop/internal/authentication"
	"avito_shop/internal/core"
	"avito_shop/internal/handlers"
	"avito_shop/internal/models"
	"avito_shop/internal/routes"
	"avito_shop/internal/storage"

	"github.com/golang-jwt/jwt"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

var (
	server *http.Server
	wg     sync.WaitGroup
)

// Настройка тестовой базы данных
func setupTestDB() (*storage.AvitoDB, *pgx.Conn, func()) {
	cfg := storage.PostgresConnConfig{
		DBHost:   "localhost",
		DBPort:   5431,
		DBName:   "shop_test",
		Username: "postgres_test",
		Password: "password",
		Options:  nil,
	}

	conn, err := storage.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	avitoDB := storage.NewAvitoDB(conn)
	authService := core.New(avitoDB)
	userHandler := handlers.NewUserHandler(authService)

	mux := routes.InitRoutes(userHandler)

	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	server = &http.Server{
		Addr:    ":8081",
		Handler: corsHandler(mux),
	}

	// Запуск сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Запуск сервера на порту 8081 http://localhost:8081/")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске сервера: %v", err)
		}
	}()

	// Возвращаем функцию для закрытия соединения с БД и остановки сервера
	return avitoDB, conn, func() {
		// Остановка сервера
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Ошибка при остановке сервера: %v", err)
		}
		wg.Wait()       // Ждем завершения работы сервера
		conn.Close(ctx) // Закрытие соединения с БД
	}
}

func addUser(employee *models.Employee, conn *pgx.Conn) error {
	//сначала хэшируем пароль
	hashedPassword, err := authentication.HashPassword(employee.Password)
	if err != nil {
		return fmt.Errorf("не удалось хэшировать пароль: %w", err)
	}
	employee.Password = hashedPassword

	// потом выдаем токен
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, authentication.GenerateTokenClaims(employee, time.Now()))
	tokenString, err := jwtToken.SignedString(authentication.SecretKey)
	if err != nil {
		return fmt.Errorf("не удалось создать токен")
	}

	employee.Token = tokenString

	err = conn.QueryRow(context.Background(), "INSERT INTO employees (username, password, token) VALUES ($1, $2, $3) RETURNING id", employee.Username, employee.Password, employee.Token).Scan(&employee.ID)
	if err != nil {
		return fmt.Errorf("ошибка при регистрации пользователя: %w", err)
	}
	return nil

}

func deleteTestData(conn *pgx.Conn) error {
	// Удаляем данные из таблицы inventory
	_, err := conn.Exec(context.Background(), "DELETE FROM inventory")
	if err != nil {
		return fmt.Errorf("ошибка при удалении данных из таблицы inventory: %w", err)
	}

	// Удаляем данные из таблицы employees
	_, err = conn.Exec(context.Background(), "DELETE FROM employees")
	if err != nil {
		return fmt.Errorf("ошибка при удалении данных из таблицы employees: %w", err)
	}

	return nil
}

func updateCoins(id int, conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "UPDATE employees SET coins = $1 WHERE id = $2", 0, id)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении данных в таблице employees: %w", err)
	}

	return nil
}

func createRequest(item string, token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8081/api/buy/%s", item), nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return req, nil
}

func Test_E2E_BuyItem_Success(t *testing.T) {
	// Запустили тестовый сервер и дб
	_, conn, teardown := setupTestDB()
	defer teardown()

	// Добавили тестового пользователя
	employee := &models.Employee{Username: "user1", Password: "password"}
	addUser(employee, conn)

	// Создаем GET-запрос на покупку товара
	req, err := createRequest("pen", employee.Token)
	require.NoError(t, err, "не удалось создать запрос")

	// Создаем сервер
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "не удалось выполнить запрос")
	defer resp.Body.Close()

	// Проверили вывод
	require.Equal(t, 204, resp.StatusCode, fmt.Sprintf("ожидался статус 200, но получил %d", resp.StatusCode))

	err = deleteTestData(conn)
	require.NoError(t, err, "не удалось удалить тестовые данные")
}

func Test_E2E_BuyItem_No_Item(t *testing.T) {
	// Запустили тестовый сервер и дб
	_, conn, teardown := setupTestDB()
	defer teardown()

	// Добавили тестового пользователя
	employee := &models.Employee{Username: "user1", Password: "password"}
	addUser(employee, conn)
	// Создаем GET-запрос на покупку товара
	req, err := createRequest("car", employee.Token)
	require.NoError(t, err, "не удалось создать запрос")

	// Создаем сервер
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "не удалось выполнить запрос")
	defer resp.Body.Close()

	// Проверили вывод
	require.Equal(t, 409, resp.StatusCode, fmt.Sprintf("ожидался статус 200, но получил %d", resp.StatusCode))

	err = deleteTestData(conn)
	require.NoError(t, err, "не удалось удалить тестовые данные")
}

func Test_E2E_BuyItem_No_Balance(t *testing.T) {
	// Запустили тестовый сервер и дб
	_, conn, teardown := setupTestDB()
	defer teardown()

	// Добавили тестового пользователя
	employee := &models.Employee{Username: "user1", Password: "password"}
	addUser(employee, conn)
	updateCoins(employee.ID, conn)

	// Создаем GET-запрос на покупку товара
	req, err := createRequest("pink-hoody", employee.Token)
	require.NoError(t, err, "не удалось создать запрос")

	// Создаем сервер
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "не удалось выполнить запрос")
	defer resp.Body.Close()

	// Проверили вывод
	require.Equal(t, 409, resp.StatusCode, fmt.Sprintf("ожидался статус 200, но получил %d", resp.StatusCode))

	err = deleteTestData(conn)
	require.NoError(t, err, "не удалось удалить тестовые данные")
}
