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

func TestHandler_Load(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name          string
		inputLogin    string
		inputOrderNum string
		expectedError error
	}{
		// определяем все тесты
		{
			name:          "Positive test for order load",
			inputLogin:    "dimma",
			inputOrderNum: "2377225624",
			expectedError: nil,
		},
	}

	s := &storagemock.Order{}
	svc := services.NewOrderService(s, nil, nil)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			err := svc.Load(ctx, tCase.inputLogin, tCase.inputOrderNum)
			// оценка результатов
			assert.Equal(t, err, err)
		})
	}
}

func TestHandler_List(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputLogin           string
		expectedStruct       []models.OrdersList
		expectedResponseBody string
		expectedError        error
	}{
		// определяем все тесты
		{
			name:       "Positive test for user wthdrawal list",
			inputLogin: "dimma",
			expectedStruct: []models.OrdersList{
				{
					Number:     "9278923470",
					Status:     "PROCESSED",
					Accrual:    decimal.NewFromFloatWithExponent(500, -2),
					UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
				{
					Number:     "12345678903",
					Status:     "PROCESSING",
					UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
				{
					Number:     "346436439",
					Status:     "INVALID",
					UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
			},
			expectedError: nil,
		},
		{
			name:       "Negative test for user wthdrawal list",
			inputLogin: "dimma2",
			expectedStruct: []models.OrdersList{
				{
					Number:     "9278923470",
					Status:     "PROCESSED",
					Accrual:    decimal.NewFromFloatWithExponent(500, -2),
					UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
				{
					Number:     "12345678903",
					Status:     "PROCESSING",
					UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
				{
					Number:     "346436439",
					Status:     "INVALID",
					UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
				},
			},
			expectedError: errors.New("something wrong with server"),
		},
	}

	s := &storagemock.Order{}
	svc := services.NewOrderService(s, nil, nil)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			ec, err := svc.List(ctx, tCase.inputLogin)
			// оценка результатов
			assert.Equal(t, tCase.expectedError, err)
			if tCase.inputLogin == "dimma" {
				assert.Equal(t, tCase.expectedStruct, ec)
			}
		})
	}
}
