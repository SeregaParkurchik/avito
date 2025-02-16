package sendcoin

import (
	"bytes"
	"context"
	"encoding/json"
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
		DBPort:   5429,
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
		Addr:    ":8082",
		Handler: corsHandler(mux),
	}

	// Запуск сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Запуск сервера на порту 8082 http://localhost:8082/")
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
	// Удаляем данные из таблицы transactions
	_, err := conn.Exec(context.Background(), "DELETE FROM transactions")
	if err != nil {
		return fmt.Errorf("ошибка при удалении данных из таблицы transactions: %w", err)
	}

	// Удаляем данные из таблицы employees
	_, err = conn.Exec(context.Background(), "DELETE FROM employees")
	if err != nil {
		return fmt.Errorf("ошибка при удалении данных из таблицы employees: %w", err)
	}

	return nil
}

func createRequest(toUser string, amount int, token string) (*http.Request, error) {
	url := "http://localhost:8082/api/sendCoin"
	data := struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}{
		ToUser: toUser,
		Amount: amount,
	}

	// Преобразуем структуру в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("не удалось преобразовать данные в JSON: %w", err)
	}

	// Создаем новый POST-запрос
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	return req, nil
}

func Test_E2E_SendCoin_Success(t *testing.T) {
	// Запустили тестовый сервер и дб
	_, conn, teardown := setupTestDB()
	defer teardown()

	// Добавили тестовых пользователей
	employee1 := &models.Employee{Username: "serega", Password: "password1"}
	addUser(employee1, conn)
	employee2 := &models.Employee{Username: "serega2", Password: "password2"}
	addUser(employee2, conn)

	// Создаем GET-запрос на покупку товара
	req, err := createRequest(employee2.Username, 50, employee1.Token) // user1 -> user2
	require.NoError(t, err, "не удалось создать запрос")

	// Создаем сервер
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "не удалось выполнить запрос")
	defer resp.Body.Close()

	// Проверили вывод
	require.Equal(t, 204, resp.StatusCode, fmt.Sprintf("ожидался статус 204, но получил %d", resp.StatusCode))

	err = deleteTestData(conn)
	require.NoError(t, err, "не удалось удалить тестовые данные")
}

func Test_E2E_SendCoin_Bad_Balance(t *testing.T) {
	// Запустили тестовый сервер и дб
	_, conn, teardown := setupTestDB()
	defer teardown()

	// Добавили тестовых пользователей
	employee1 := &models.Employee{Username: "serega", Password: "password1"}
	addUser(employee1, conn)
	employee2 := &models.Employee{Username: "dima", Password: "password2"}
	addUser(employee2, conn)

	// Создаем GET-запрос на покупку товара
	req, err := createRequest(employee2.Username, 1001, employee1.Token) // user1 -> user2
	require.NoError(t, err, "не удалось создать запрос")

	// Создаем сервер
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "не удалось выполнить запрос")
	defer resp.Body.Close()

	// Проверили вывод
	require.Equal(t, 409, resp.StatusCode, fmt.Sprintf("ожидался статус 409, но получил %d", resp.StatusCode))

	err = deleteTestData(conn)
	require.NoError(t, err, "не удалось удалить тестовые данные")
}

func Test_E2E_SendCoin_No_Recipient(t *testing.T) {
	// Запустили тестовый сервер и дб
	_, conn, teardown := setupTestDB()
	defer teardown()

	// Добавили тестовых пользователей
	employee1 := &models.Employee{Username: "serega", Password: "password1"}
	addUser(employee1, conn)
	employee2 := &models.Employee{Username: "dima", Password: "password2"}
	addUser(employee2, conn)

	// Создаем GET-запрос на покупку товара
	req, err := createRequest("user2", 1001, employee1.Token) // user1 -> user2
	require.NoError(t, err, "не удалось создать запрос")

	// Создаем сервер
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "не удалось выполнить запрос")
	defer resp.Body.Close()

	// Проверили вывод
	require.Equal(t, 409, resp.StatusCode, fmt.Sprintf("ожидался статус 409, но получил %d", resp.StatusCode))

	err = deleteTestData(conn)
	require.NoError(t, err, "не удалось удалить тестовые данные")
}
