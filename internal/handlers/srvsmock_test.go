package handlers

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers/mocks"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHandler_HandlerNewUserReg(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputMetod           string
		inputEndpoint        string
		inputBody            string
		dc                   models.DecodeLoginPair
		expectedStatusCode   int
		expectedResponseBody string
		expectedContentType  string
	}{
		// определяем все тесты
		{
			name:          "positive test user registration",
			inputMetod:    "POST",
			inputEndpoint: "/api/user/register",
			inputBody: `{
					"login": "dimma",
					"password": "12345"
				}`,
			dc:                   models.DecodeLoginPair{Login: "dimma", Password: "12345"},
			expectedStatusCode:   200,
			expectedResponseBody: "Authorization",
			expectedContentType:  "text/plain; charset=utf-8",
		},
	}
	s := &mocks.MockService{}
	h := NewHandler(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// Test Server
			rout := chi.NewRouter()
			rout.Post("/api/user/register", h.HandlerNewUserReg)
			//запрос
			request := httptest.NewRequest(tCase.inputMetod, tCase.inputEndpoint, bytes.NewBufferString(tCase.inputBody))
			// create Test Request
			w := httptest.NewRecorder()
			// perform Test Request
			rout.ServeHTTP(w, request)
			// Assert results
			assert.Equal(t, tCase.expectedStatusCode, w.Code)
			assert.Contains(t, tCase.expectedResponseBody, w.Body.String())
		})
	}
}
