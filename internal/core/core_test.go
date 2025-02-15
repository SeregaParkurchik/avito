package core

import (
	"avito_shop/internal/authentication"
	"avito_shop/internal/models"
	"avito_shop/internal/storage"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Auth(t *testing.T) {
	t.Parallel()

	// Arrange
	validEmployee := &models.Employee{
		ID:       0,
		Username: "serega11111",
		Password: "password",
	}

	type testCase struct {
		name             string
		expectedEmployee *models.Employee
		expectedResult   string
		needError        bool
		mockSetup        func(db *storage.MockInterface) *storage.MockInterface
		now              time.Time
	}

	testCases := []testCase{
		{
			name:             "CheckEmployeeError",
			expectedEmployee: &models.Employee{Username: "serega11111"},
			expectedResult:   "",
			needError:        true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().CheckEmployee("serega11111").Return(false, errors.New("нет пользователя"))

				return db
			},
		},
		{
			name:             "EmptyUsernameError",
			expectedEmployee: &models.Employee{Username: ""},
			expectedResult:   "",
			needError:        true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				return db
			},
		},
		{
			name: "GenerateHashError",
			expectedEmployee: &models.Employee{
				Username: "serega11111",
				Password: string(make([]byte, 75)),
			},
			expectedResult: "",
			needError:      true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().CheckEmployee("serega11111").Return(false, nil)

				return db
			},
		},
		{
			name:             "RegisterError",
			expectedEmployee: validEmployee,
			expectedResult:   "",
			needError:        true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().CheckEmployee("serega11111").Return(false, nil)
				db.EXPECT().Register(validEmployee).Return(errors.New("error"))

				return db
			},
		},
		{
			name:             "RegisterSuccess",
			expectedEmployee: validEmployee,
			expectedResult:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoic2VyZWdhMTExMTEifSwiaWF0IjowLCJleHAiOjQzMjAwfQ.Ywy-K0Cq6PcRcHY9wOh2PbmUgeI9uvU7ABbMwg7som4",
			needError:        false,
			now:              time.UnixMicro(10),
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().CheckEmployee("serega11111").Return(false, nil)
				db.EXPECT().Register(validEmployee).Return(nil)

				return db
			},
		},
		{
			name:             "LoginError",
			expectedEmployee: validEmployee,
			expectedResult:   "",
			needError:        true,
			now:              time.UnixMicro(10),
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().CheckEmployee("serega11111").Return(true, nil)
				db.EXPECT().Login(validEmployee).Return(models.Employee{}, errors.New("error"))

				return db
			},
		},
		{
			name:             "LoginSuccess",
			expectedEmployee: validEmployee,
			expectedResult:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoiIn0sImlhdCI6MCwiZXhwIjo0MzIwMH0.s_cWazHgKuENaY0HUaCUfecOpMsdjNYQ8jpcGz705wA",
			needError:        false,
			now:              time.UnixMicro(10),
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().CheckEmployee("serega11111").Return(true, nil)
				db.EXPECT().Login(validEmployee).Return(models.Employee{Password: "$2a$10$PHxvWXwshbIKhhPCVh.hfeXPy8T0eOP7Hxk6id0/U6oc256aZorAq"}, nil)
				db.EXPECT().UpdateToken(0, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoiIn0sImlhdCI6MCwiZXhwIjo0MzIwMH0.s_cWazHgKuENaY0HUaCUfecOpMsdjNYQ8jpcGz705wA").Return(nil)

				return db
			},
		},
		{
			name:             "CheckHashError",
			expectedEmployee: validEmployee,
			expectedResult:   "",
			needError:        true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().CheckEmployee("serega11111").Return(true, nil)
				db.EXPECT().Login(validEmployee).Return(models.Employee{Password: "password1"}, nil)

				return db
			},
		},
		// {
		// 	name:             "UpdateTokenError",
		// 	expectedEmployee: validEmployee,
		// 	expectedResult:   "",
		// 	needError:        true,
		// 	mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
		// 		db.EXPECT().CheckEmployee("serega11111").Return(true, nil)
		// 		//db.EXPECT().Login(validEmployee).Return(models.Employee{Password: "password"}, nil)
		// 		//db.EXPECT().CheckPasswordHash()
		// 		db.EXPECT().UpdateToken(validEmployee.ID, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoiIn0sImlhdCI6MCwiZXhwIjo0MzIwMH0.s_cWazHgKuENaY0HUaCUfecOpMsdjNYQ8jpcGz705wA").Return(errors.New("error"))
		// 		return db
		// 	},
		// },
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			dbMock := tt.mockSetup(storage.NewMockInterface(t))
			serv := &service{
				storage: dbMock,
			}

			// Act
			token, err := serv.Auth(context.Background(), tt.expectedEmployee, tt.now)

			// Assert

			if tt.needError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, token)
		})
	}

}

func Test_BuyItem(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		item      string
		username  string
		needError bool
		mockSetup func(db *storage.MockInterface) *storage.MockInterface
	}

	testCases := []testCase{
		{
			name:      "SuccessfulBuyItem",
			item:      "pen",
			username:  "serega",
			needError: false,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.On("BuyItem", "pen", "serega").Return(nil)
				return db
			},
		},
		{
			name:      "WrongItem",
			item:      "item123",
			username:  "serega",
			needError: true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.On("BuyItem", "item123", "serega").Return(errors.New("товара не существует"))
				return db
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			dbMock := tt.mockSetup(storage.NewMockInterface(t))
			serv := &service{
				storage: dbMock,
			}

			// Act
			err := serv.BuyItem(context.Background(), tt.item, tt.username)

			// Assert
			if tt.needError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_SendCoin(t *testing.T) {
	t.Parallel()

	// Arrange
	validSend := &models.SendCoin{
		FromUser: "user1",
		ToUser:   "user2",
		Amount:   10,
	}

	invalidSend := &models.SendCoin{
		FromUser: "user1",
		ToUser:   "user1",
		Amount:   10,
	}

	insufficientFundsSend := &models.SendCoin{
		FromUser: "user1",
		ToUser:   "user2",
		Amount:   1000, // Предположим, что у пользователя недостаточно средств
	}

	type testCase struct {
		name      string
		send      *models.SendCoin
		username  string
		needError bool
		mockSetup func(db *storage.MockInterface) *storage.MockInterface
	}

	testCases := []testCase{
		{
			name:      "SuccessSend",
			send:      validSend,
			username:  "user1",
			needError: false,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().SendCoin(validSend).Return(nil)
				return db
			},
		},
		{
			name:      "SendToSelf",
			send:      invalidSend,
			username:  "user1",
			needError: true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				return db
			},
		},
		{
			name:      "InsufficientFunds",
			send:      insufficientFundsSend,
			username:  "user1",
			needError: true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().SendCoin(insufficientFundsSend).Return(fmt.Errorf("недостаточно средств"))
				return db
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			dbMock := tt.mockSetup(storage.NewMockInterface(t))
			serv := &service{
				storage: dbMock,
			}

			// Act
			err := serv.SendCoin(context.Background(), tt.send, tt.username)

			// Assert
			if tt.needError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_Info(t *testing.T) {
	t.Parallel()

	// Arrange
	username := "user1"

	validCoins := 100
	validItems := []models.Item{
		{Type: "pen", Quantity: 2},
		{Type: "cup", Quantity: 1},
	}
	validSent := []models.SentTransaction{
		{FromUser: "user1", Amount: 20},
		{FromUser: "user2", Amount: 30},
	}
	validReceived := []models.ReceivedTransaction{
		{ToUser: "user2", Amount: 15},
		{ToUser: "user1", Amount: 5},
	}

	type testCase struct {
		name      string
		username  string
		expected  models.InfoResponse
		needError bool
		mockSetup func(db *storage.MockInterface) *storage.MockInterface
	}

	testCases := []testCase{
		{
			name:     "SuccessInfo",
			username: username,
			expected: models.InfoResponse{
				Coins:     validCoins,
				Inventory: validItems,
				CoinHistory: models.CoinHistory{
					Sent:     validSent,
					Received: validReceived,
				},
			},
			needError: false,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().GetCoins(username).Return(validCoins, nil)
				db.EXPECT().GetInventory(username).Return(validItems, nil)
				db.EXPECT().GetTransaction(username).Return(validSent, validReceived, nil)
				return db
			},
		},
		{
			name:      "ErrorGettingCoins",
			username:  username,
			expected:  models.InfoResponse{},
			needError: true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().GetCoins(username).Return(0, errors.New("error getting coins"))
				return db
			},
		},
		{
			name:      "ErrorGettingInventory",
			username:  username,
			expected:  models.InfoResponse{},
			needError: true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().GetCoins(username).Return(validCoins, nil)
				db.EXPECT().GetInventory(username).Return(nil, errors.New("error getting inventory"))
				return db
			},
		},
		{
			name:      "ErrorGettingTransaction",
			username:  username,
			expected:  models.InfoResponse{},
			needError: true,
			mockSetup: func(db *storage.MockInterface) *storage.MockInterface {
				db.EXPECT().GetCoins(username).Return(validCoins, nil)
				db.EXPECT().GetInventory(username).Return(validItems, nil)
				db.EXPECT().GetTransaction(username).Return(nil, nil, errors.New("error getting transaction"))
				return db
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			dbMock := tt.mockSetup(storage.NewMockInterface(t))
			serv := &service{
				storage: dbMock,
			}

			// Act
			response, err := serv.Info(context.Background(), tt.username)

			// Assert
			if tt.needError {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, response)
			}
		})
	}
}

func Test_Test(t *testing.T) {
	hash, err := authentication.HashPassword("password")
	require.NoError(t, err)

	t.Logf("hash: %s", hash)

	validEmployee := &models.Employee{
		Username: "serega11111",
		Password: "$2a$10$eaE5wSBL3unWmzQ2E9dEu.s9YAcPxKFx1K5h4ihx4C46OxstKXgv.",
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, authentication.GenerateTokenClaims(validEmployee, time.UnixMicro(10)))
	tokenString, err := jwtToken.SignedString(authentication.SecretKey)
	require.NoError(t, err)

	t.Logf("tokenString: %s", tokenString)
}
