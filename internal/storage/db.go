//go:generate mockery --filename db_mock.go --name Interface --inpackage --with-expecter
package storage

import (
	"avito_shop/internal/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Interface interface {
	CheckEmployee(username string) (bool, error)
	Register(employee *models.Employee) error
	Login(employee *models.Employee) (models.Employee, error)
	UpdateToken(id int, token string) error
	BuyItem(item string, username string) error
	SendCoin(send *models.SendCoin) error
	GetCoins(username string) (int, error)
	GetInventory(username string) ([]models.Item, error)
	GetTransaction(username string) ([]models.SentTransaction, []models.ReceivedTransaction, error)
}

type PostgresConnConfig struct {
	DBHost   string
	DBPort   uint
	DBName   string
	Username string
	Password string
	Options  map[string]string
}

type AvitoDB struct {
	conn *pgx.Conn
}

func NewAvitoDB(conn *pgx.Conn) *AvitoDB {
	return &AvitoDB{conn: conn}
}

func New(ctx context.Context, cfg PostgresConnConfig) (*pgx.Conn, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Username, cfg.Password, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	var options string
	if len(cfg.Options) > 0 {
		for key, value := range cfg.Options {
			options += fmt.Sprintf("%s=%s&", key, value)
		}

		options = options[:len(options)-1]
		connStr += options
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to postgres: %w", err)
	}

	return conn, nil
}

func (s *AvitoDB) CheckEmployee(username string) (bool, error) {
	var exists bool
	err := s.conn.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM employees WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}
	return exists, nil
}

func (s *AvitoDB) Register(employee *models.Employee) error {
	err := s.conn.QueryRow(context.Background(), "INSERT INTO employees (username, password, token) VALUES ($1, $2, $3) RETURNING id", employee.Username, employee.Password, employee.Token).Scan(&employee.ID)
	if err != nil {
		return fmt.Errorf("ошибка при регистрации пользователя: %w", err)
	}
	return nil
}

func (s *AvitoDB) Login(employee *models.Employee) (models.Employee, error) {
	var foundEmployee models.Employee
	err := s.conn.QueryRow(context.Background(), "SELECT id, username, password, coins FROM employees WHERE username = $1", employee.Username).Scan(&foundEmployee.ID, &foundEmployee.Username, &foundEmployee.Password, &foundEmployee.Coins)
	if err != nil {
		return models.Employee{}, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}

	return foundEmployee, nil
}

func (s *AvitoDB) UpdateToken(id int, token string) error {
	_, err := s.conn.Exec(context.Background(), "UPDATE employees SET token = $1 WHERE id = $2", token, id)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении токена: %w", err)
	}
	return nil
}

func (s *AvitoDB) BuyItem(item string, username string) error {
	var priceItem int
	var employeeID int
	var merchandiseID int
	// Получаем цену товара
	err := s.conn.QueryRow(context.Background(), "SELECT price FROM merchandise WHERE name = $1", item).Scan(&priceItem)
	if err != nil {
		return fmt.Errorf("товара не существует: %w", err)
	}

	// Получаем ID пользователя
	err = s.conn.QueryRow(context.Background(), "SELECT id FROM employees WHERE username = $1", username).Scan(&employeeID)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	// Получаем ID товара
	err = s.conn.QueryRow(context.Background(), "SELECT id FROM merchandise WHERE name = $1", item).Scan(&merchandiseID)
	if err != nil {
		return fmt.Errorf("товар не найден: %w", err)
	}

	// Обновляем количество монет у пользователя
	result, err := s.conn.Exec(context.Background(), `
		UPDATE employees 
		SET coins = coins - $1 
		WHERE id = $2 AND coins >= $1`, priceItem, employeeID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении монет: %w", err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("недостаточно монет для покупки товара")
	}

	// Добавляем товар в инвентарь
	_, err = s.conn.Exec(context.Background(), `
		INSERT INTO inventory (employee_id, merchandise_id, quantity)
		VALUES ($1, $2, 1)
		ON CONFLICT (employee_id, merchandise_id) 
		DO UPDATE SET quantity = inventory.quantity + 1`, employeeID, merchandiseID)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении товара в инвентарь: %w", err)
	}

	return nil
}

func (s *AvitoDB) SendCoin(send *models.SendCoin) error {
	var senderCoins int
	var senderID int
	var receiverID int

	// Получаем баланс отправителя
	err := s.conn.QueryRow(context.Background(), "SELECT coins, id FROM employees WHERE username = $1", send.FromUser).Scan(&senderCoins, &senderID)
	if err != nil {
		return fmt.Errorf("пользователь-отправитель не найден: %w", err)
	}

	// Проверяем, достаточно ли монет у отправителя
	if senderCoins < send.Amount {
		return fmt.Errorf("недостаточно монет у отправителя %s", send.FromUser)
	}

	// Получаем ID получателя
	err = s.conn.QueryRow(context.Background(), "SELECT id FROM employees WHERE username = $1", send.ToUser).Scan(&receiverID)
	if err != nil {
		return fmt.Errorf("пользователь-получатель не найден: %w", err)
	}

	// Обновляем количество монет у отправителя
	_, err = s.conn.Exec(context.Background(), `
        UPDATE employees 
        SET coins = coins - $1 
        WHERE id = $2`, send.Amount, senderID)
	if err != nil {
		return fmt.Errorf("ошибка при списании монет у отправителя: %w", err)
	}

	// Начисляем монеты получателю
	_, err = s.conn.Exec(context.Background(), `
        UPDATE employees 
        SET coins = coins + $1 
        WHERE id = $2`, send.Amount, receiverID)
	if err != nil {
		return fmt.Errorf("ошибка при начислении монет получателю: %w", err)
	}

	// Записываем историю транзакций
	_, err = s.conn.Exec(context.Background(), `
        INSERT INTO transactions (employee_id, transaction_type, amount, from_user, to_user) 
        VALUES ($1, 'sent', $2, $3, NULL)`, senderID, send.Amount, send.FromUser)
	if err != nil {
		return fmt.Errorf("ошибка при записи транзакции отправителя: %w", err)
	}

	_, err = s.conn.Exec(context.Background(), `
        INSERT INTO transactions (employee_id, transaction_type, amount, from_user, to_user) 
        VALUES ($1, 'received', $2, NULL, $3)`, receiverID, send.Amount, send.ToUser)
	if err != nil {
		return fmt.Errorf("ошибка при записи транзакции получателя: %w", err)
	}

	return nil
}

func (s *AvitoDB) GetCoins(username string) (int, error) {
	var coins int

	err := s.conn.QueryRow(context.Background(), "SELECT coins FROM employees WHERE username = $1", username).Scan(&coins)
	if err != nil {
		return 0, fmt.Errorf("не удалось получить монеты: %w", err)
	}

	return coins, nil
}

func (s *AvitoDB) GetInventory(username string) ([]models.Item, error) {
	var items []models.Item

	// Выполняем SQL-запрос для получения данных из инвентаря
	rows, err := s.conn.Query(context.Background(), `
        SELECT m.name AS type, i.quantity 
        FROM inventory i
        JOIN employees e ON i.employee_id = e.id
        JOIN merchandise m ON i.merchandise_id = m.id
        WHERE e.username = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()

	// Проходим по результатам запроса
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			return nil, fmt.Errorf("ошибка при считывании строки: %w", err)
		}
		items = append(items, item)
	}

	// Проверяем на наличие ошибок после итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return items, nil
}

func (s *AvitoDB) GetTransaction(username string) ([]models.SentTransaction, []models.ReceivedTransaction, error) {
	var employeesID int
	// Получили айди
	err := s.conn.QueryRow(context.Background(), "SELECT id FROM employees WHERE username = $1", username).Scan(&employeesID)
	if err != nil {
		return nil, nil, fmt.Errorf("пользователь-получатель не найден: %w", err)
	}

	// Запрос для получения всех транзакций типа 'sent'
	rows, err := s.conn.Query(context.Background(), "SELECT amount, from_user FROM transactions WHERE transaction_type = 'sent' AND employee_id = $1", employeesID)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка при получении транзакций: %w", err)
	}
	defer rows.Close()

	var sentTransactions []models.SentTransaction

	// Обработка результатов запроса
	for rows.Next() {
		var t models.SentTransaction

		err := rows.Scan(&t.Amount, &t.FromUser)
		if err != nil {
			return nil, nil, fmt.Errorf("ошибка при сканировании строки: %w", err)
		}

		sentTransactions = append(sentTransactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	// Запрос для получения всех транзакций типа 'received'
	receivedRows, err := s.conn.Query(context.Background(), "SELECT amount, to_user FROM transactions WHERE transaction_type = 'received' AND employee_id = $1", employeesID)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка при получении полученных транзакций: %w", err)
	}
	defer receivedRows.Close()

	var receivedTransactions []models.ReceivedTransaction

	// Обработка результатов запроса для 'received'
	for receivedRows.Next() {
		var t models.ReceivedTransaction

		err := receivedRows.Scan(&t.Amount, &t.ToUser)
		if err != nil {
			return nil, nil, fmt.Errorf("ошибка при сканировании строки полученных транзакций: %w", err)
		}

		receivedTransactions = append(receivedTransactions, t)
	}

	if err := receivedRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("ошибка при итерации по строкам полученных транзакций: %w", err)
	}

	return sentTransactions, receivedTransactions, nil
}
