package service__test

import (
	"context"
	"errors"
	"testing"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/shopspring/decimal"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services/storagemock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Create(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputLogin           string
		expectedStruct       models.DecodeLoginPair
		expectedResponseBody string
		expectedError        error
	}{
		// определяем все тесты
		{
			name:       "Positive test for user balance status",
			inputLogin: "dimma",
			expectedStruct: models.LoginBalance{
				Current:   decimal.NewFromFloatWithExponent(500.505, -2),
				Withdrawn: decimal.NewFromFloatWithExponent(42, -2),
			},
			expectedError: errors.New("something wrong woth server"),
		},
	}

	s := &storagemock.Balance{}
	svc := services.NewBalanceService(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			ec, err := svc.Status(ctx, tCase.inputLogin)
			// оценка результатов
			assert.Equal(t, tCase.expectedError, err)
			assert.Equal(t, tCase.expectedStruct, ec)

		})
	}
}

func TestHandler_CheckAuthorization(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputLogin           string
		inputStruct          models.NewWithdrawal
		expectedResponseBody string
		expectedError        error
	}{
		// определяем все тесты
		{
			name:       "Positive test for user  new wthdrawal",
			inputLogin: "dimma",
			inputStruct: models.NewWithdrawal{
				Order: "2377225624",
				Sum:   decimal.NewFromFloatWithExponent(42, -2),
			},
			expectedError: nil,
		},
		{
			name:       "Negative test for user  new wthdrawal",
			inputLogin: "dimma",
			inputStruct: models.NewWithdrawal{
				Order: "",
				Sum:   decimal.NewFromFloatWithExponent(42, -2),
			},
			expectedError: errors.New("something wrong woth server"),
		},
	}

	s := &storagemock.Balance{}
	svc := services.NewBalanceService(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			err := svc.NewWithdrawal(ctx, tCase.inputLogin, tCase.inputStruct)
			// оценка результатов
			assert.Equal(t, tCase.expectedError, err)
		})
	}
}
