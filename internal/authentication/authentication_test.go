package authentication

import (
	"avito_shop/internal/models"
	"testing"
	"time"
)

func Test_HashPassword(t *testing.T) {
	t.Parallel()

	validPass := "password"

	type testCase struct {
		name           string
		password       string
		expectedResult string
		needError      bool
	}

	testCases := []testCase{
		{
			name:           "HashPasswordSuccess",
			password:       validPass,
			expectedResult: "$2a$10$Kx/YB1Kqb3E8/BRa7ELPGub9cy17RUmwjor5xFcj9kgSQGbTgWsZa",
			needError:      false,
		},
		{
			name:           "HashPasswordError",
			password:       string(make([]byte, 75)),
			expectedResult: "",
			needError:      true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := HashPassword(tc.password)

			if tc.needError {
				if err == nil {
					t.Errorf("ожидалась ошибка, но ее не было")
				}
				return
			}

			if err != nil {
				t.Errorf("ожидалась отсутствие ошибки, но получена: %v", err)
			}
		})
	}
}

func Test_CheckPasswordHash(t *testing.T) {
	t.Parallel()

	validPass := "password"
	hashedPass := "$2a$10$Kx/YB1Kqb3E8/BRa7ELPGub9cy17RUmwjor5xFcj9kgSQGbTgWsZa"

	type testCase struct {
		name           string
		password       string
		hash           string
		expectedResult bool
	}

	testCases := []testCase{
		{
			name:           "CheckPasswordHashSuccess",
			password:       validPass,
			hash:           hashedPass,
			expectedResult: true,
		},
		{
			name:           "CheckPasswordHashFailure",
			password:       "wrongpassword",
			hash:           hashedPass,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CheckPasswordHash(tc.password, tc.hash)

			if result != tc.expectedResult {
				t.Errorf("ожидался результат %v, но получен: %v", tc.expectedResult, result)
			}
		})
	}
}

func Test_Valid(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		claims        TokenClaims
		expectError   bool
		expectedError string
	}

	currentTime := time.Now().Unix()

	testCases := []testCase{
		{
			name:        "ValidToken",
			claims:      TokenClaims{EXP: currentTime + 3600}, // Токен действителен на 1 час вперед
			expectError: false,
		},
		{
			name:          "ExpiredToken",
			claims:        TokenClaims{EXP: currentTime - 1}, // Токен истек
			expectError:   true,
			expectedError: "токен истек",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.claims.Valid()

			if tc.expectError {
				if err == nil {
					t.Errorf("ожидалась ошибка, но ее не было")
				} else if err.Error() != tc.expectedError {
					t.Errorf("ожидалась ошибка '%v', но получена: '%v'", tc.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("ожидалась отсутствие ошибки, но получена: %v", err)
				}
			}
		})
	}
}

func Test_GenerateTokenClaims(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		employee    models.Employee
		now         time.Time
		expectedIAT int64
		expectedEXP int64
	}

	testCases := []testCase{
		{
			name: "ValidClaims",
			employee: models.Employee{
				Username: "testuser",
			},
			now:         time.Unix(1609459200, 0), // 1 января 2021 года
			expectedIAT: 1609459200,               // IAT должен совпадать с now
			expectedEXP: 1609502400,               // EXP должен быть через 12 часов
		},
		{
			name: "AnotherUser ",
			employee: models.Employee{
				Username: "anotheruser",
			},
			now:         time.Unix(1612137600, 0), // 1 февраля 2021 года
			expectedIAT: 1612137600,               // IAT должен совпадать с now
			expectedEXP: 1612180800,               // EXP должен быть через 12 часов
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims := GenerateTokenClaims(&tc.employee, tc.now)

			if claims.IAT != tc.expectedIAT {
				t.Errorf("ожидалось IAT '%v', но получено: '%v'", tc.expectedIAT, claims.IAT)
			}
			if claims.EXP != tc.expectedEXP {
				t.Errorf("ожидалось EXP '%v', но получено: '%v'", tc.expectedEXP, claims.EXP)
			}
			if claims.Employee.Username != tc.employee.Username {
				t.Errorf("ожидалось Username '%v', но получено: '%v'", tc.employee.Username, claims.Employee.Username)
			}
		})
	}
}
