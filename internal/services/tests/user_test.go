// тесты слоя сервисов
package service__test

import (
	"context"
	"testing"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services/storagemock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Create(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputLogin           string
		inputStruct          models.DecodeLoginPair
		expectedResponseBody string
		expectedError        error
	}{
		// определяем все тесты
		{
			name:       "Positive test for user login create",
			inputLogin: "dimma",
			inputStruct: models.DecodeLoginPair{
				Login:    "dimma",
				Password: "12345",
			},
			expectedError: nil,
		},
	}

	s := &storagemock.User{}
	svc := services.NewUserService(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			err := svc.Create(ctx, tCase.inputStruct)
			// оценка результатов
			assert.Equal(t, tCase.expectedError, err)
		})
	}
}

func TestHandler_CheckAuthorization(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputLogin           string
		inputStruct          models.DecodeLoginPair
		expectedResponseBody string
		expectedError        error
	}{
		// определяем все тесты
		{
			name:       "Positive test for user CheckAuthorization(",
			inputLogin: "dimma",
			inputStruct: models.DecodeLoginPair{
				Login:    "dimma",
				Password: "12345",
			},
			expectedError: nil,
		},
	}

	s := &storagemock.User{}
	svc := services.NewUserService(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			err := svc.CheckAuthorization(ctx, tCase.inputStruct)
			// оценка результатов
			assert.Equal(t, tCase.expectedError, err)
		})
	}
}
