//go:generate mockery --filename core_mock.go --name Interface --inpackage --with-expecter
package core

import (
	"avito_shop/internal/authentication"
	"avito_shop/internal/models"
	"avito_shop/internal/storage"
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type Interface interface {
	Auth(ctx context.Context, employee *models.Employee, now time.Time) (string, error)
	BuyItem(ctx context.Context, item string, username string) error
	SendCoin(ctx context.Context, send *models.SendCoin, username string) error
	Info(ctx context.Context, username string) (models.InfoResponse, error)
}

type service struct {
	storage storage.Interface
}

func New(storage storage.Interface) Interface {
	return &service{
		storage: storage,
	}
}

func (s *service) Auth(ctx context.Context, employee *models.Employee, now time.Time) (string, error) {
	if employee.Username == "" {
		return "", fmt.Errorf("неправильные входные данные")
	}

	ok, err := s.storage.CheckEmployee(employee.Username)
	if err != nil {
		return "", err
	}

	if !ok {
		//сначала хэшируем пароль
		hashedPassword, err := authentication.HashPassword(employee.Password)
		if err != nil {
			return "", fmt.Errorf("не удалось хэшировать пароль: %w", err)
		}
		employee.Password = hashedPassword

		// потом выдаем токен
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, authentication.GenerateTokenClaims(employee, now))
		tokenString, err := jwtToken.SignedString(authentication.SecretKey)
		if err != nil {
			return "", fmt.Errorf("не удалось создать токен")
		}

		employee.Token = tokenString

		err = s.storage.Register(employee)
		if err != nil {
			return "", err
		}

		return tokenString, nil
	}

	foundEmployee, err := s.storage.Login(employee)
	if err != nil {
		return "", err
	}

	if !authentication.CheckPasswordHash(employee.Password, foundEmployee.Password) {
		return "", fmt.Errorf("неверный пароль")
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, authentication.GenerateTokenClaims(&foundEmployee, now))

	tokenString, err := jwtToken.SignedString(authentication.SecretKey)
	if err != nil {
		return "", fmt.Errorf("не удалось создать токен")
	}

	err = s.storage.UpdateToken(foundEmployee.ID, tokenString)
	if err != nil {
		return "", fmt.Errorf("не удалось обновить токен")
	}

	return tokenString, nil
}

func (s *service) BuyItem(ctx context.Context, item string, username string) error {
	err := s.storage.BuyItem(item, username)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) SendCoin(ctx context.Context, send *models.SendCoin, username string) error {
	send.FromUser = username

	if send.ToUser == send.FromUser || send.Amount < 0 {
		return fmt.Errorf("самому себе отправить монеты нельзя, нельзя отправить отрицательное значение")
	}

	err := s.storage.SendCoin(send)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Info(ctx context.Context, username string) (models.InfoResponse, error) {
	coins, err := s.storage.GetCoins(username)
	if err != nil {
		return models.InfoResponse{}, err
	}

	items, err := s.storage.GetInventory(username)
	if err != nil {
		return models.InfoResponse{}, err
	}

	sent, received, err := s.storage.GetTransaction(username)
	if err != nil {
		return models.InfoResponse{}, err
	}

	var response models.InfoResponse

	response.Coins = coins
	response.Inventory = items

	coinHistory := models.CoinHistory{
		Received: received,
		Sent:     sent,
	}
	response.CoinHistory = coinHistory

	return response, nil
}
