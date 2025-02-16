package handlers

import (
	"avito_shop/internal/authentication"
	"avito_shop/internal/core"
	"avito_shop/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_UserHandler_Auth(t *testing.T) {
	t.Parallel()

	// Arrange
	validEmployee := &models.Employee{
		ID:       0,
		Username: "serega11111",
		Password: "password",
	}

	type authTestCase struct {
		name          string
		requestBody   *models.Employee
		expectedToken string
		expectError   bool
		mockSetup     func(core *core.MockInterface)
		now           time.Time
	}

	testCases := []authTestCase{
		{
			name:          "AuthError",
			requestBody:   validEmployee,
			expectedToken: "",
			expectError:   true,
			now:           time.UnixMicro(10),
			mockSetup: func(core *core.MockInterface) {
				core.EXPECT().Auth(context.Background(), validEmployee, mock.Anything).Return("", errors.New("authentication failed"))

			},
		},
		{
			name:          "AuthSuccess",
			requestBody:   validEmployee,
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoic2VyZWdhMTExMTEifSwiaWF0IjowLCJleHAiOjQzMjAwfQ.Ywy-K0Cq6PcRcHY9wOh2PbmUgeI9uvU7ABbMwg7som4",
			expectError:   false,
			now:           time.UnixMicro(10),
			mockSetup: func(core *core.MockInterface) {
				core.EXPECT().Auth(context.Background(), validEmployee, mock.Anything).Return("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoic2VyZWdhMTExMTEifSwiaWF0IjowLCJleHAiOjQzMjAwfQ.Ywy-K0Cq6PcRcHY9wOh2PbmUgeI9uvU7ABbMwg7som4", nil)

			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый мок-сервис
			coreMock := core.NewMockInterface(t)
			// Настраиваем мок-сервис
			tt.mockSetup(coreMock)

			// Создаем новый экземпляр UserHandler с мок-сервисом
			handler := &UserHandler{service: coreMock}

			// Создаем новый HTTP-запрос
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewBuffer(requestBody))
			w := httptest.NewRecorder()

			// Вызываем обработчик
			handler.Auth(w, req)

			// Проверяем ответ
			res := w.Result()
			if tt.expectError {
				assert.Equal(t, http.StatusConflict, res.StatusCode)
				body, _ := io.ReadAll(res.Body)
				assert.Contains(t, string(body), "authentication failed")
			} else {
				assert.Equal(t, http.StatusOK, res.StatusCode)
				var response authentication.RegisterResponse
				json.NewDecoder(res.Body).Decode(&response)
				assert.Equal(t, tt.expectedToken, response.AccessToken)
			}
		})
	}
}

func Test_UserHandler_BuyItem(t *testing.T) {
	t.Parallel()

	type buyItemTestCase struct {
		name         string
		item         string
		username     string
		expectError  bool
		expectedCode int
		mockSetup    func(core *core.MockInterface)
	}

	testCases := []buyItemTestCase{
		{
			name:         "UserNotAuthenticated",
			item:         "item1",
			username:     "",
			expectError:  true,
			expectedCode: http.StatusUnauthorized,
			mockSetup: func(core *core.MockInterface) {
				// Здесь мок не требуется, так как ошибка возникает до вызова сервиса
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый мок-сервис
			coreMock := core.NewMockInterface(t)
			// Настраиваем мок-сервис
			tt.mockSetup(coreMock)

			// Создаем новый экземпляр UserHandler с мок-сервисом
			handler := &UserHandler{service: coreMock}

			// Создаем новый HTTP-запрос с контекстом
			req := httptest.NewRequest(http.MethodPost, "/buy/"+tt.item, nil)
			if tt.username != "" {
				ctx := context.WithValue(req.Context(), usernameKey, tt.username)
				req = req.WithContext(ctx)
			}
			w := httptest.NewRecorder()

			// Вызываем обработчик
			handler.BuyItem(w, req)

			// Проверяем ответ
			res := w.Result()
			assert.Equal(t, tt.expectedCode, res.StatusCode)
			if tt.expectError {
				body, _ := io.ReadAll(res.Body)
				assert.Contains(t, string(body), "не удалось извлечь имя пользователя из контекста\n")
			}
		})
	}
}
