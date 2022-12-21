package service__test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/shopspring/decimal"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services/storagemock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Status(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputLogin           string
		expectedStruct       models.LoginBalance
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

func TestHandler_NewWithdrawal(t *testing.T) {
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

func TestHandler_WithdrawalsList(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputLogin           string
		inputStruct          []models.WithdrawalsList
		expectedResponseBody string
		expectedError        error
	}{
		// определяем все тесты
		{
			name:       "Positive test for user wthdrawal list",
			inputLogin: "dimma",
			inputStruct: []models.WithdrawalsList{
				{
					Order:       "2377225624",
					Sum:         decimal.NewFromFloatWithExponent(500.0300, -2),
					ProcessedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
				{
					Order:       "2377225625",
					Sum:         decimal.NewFromFloatWithExponent(800.5555, -2),
					ProcessedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
			},
			expectedError: nil,
		},
		{
			name:       "Negative test for user wthdrawal list",
			inputLogin: "dimma2",
			inputStruct: []models.WithdrawalsList{
				{
					Order:       "2377225624",
					Sum:         decimal.NewFromFloatWithExponent(500.0300, -2),
					ProcessedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
				{
					Order:       "2377225625",
					Sum:         decimal.NewFromFloatWithExponent(800.5555, -2),
					ProcessedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
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
			ec, err := svc.WithdrawalsList(ctx, tCase.inputLogin)
			// оценка результатов
			assert.Equal(t, tCase.expectedError, err)
			if tCase.inputLogin == "dimma" {
				assert.Equal(t, tCase.inputStruct, ec)
			}
		})
	}
}
