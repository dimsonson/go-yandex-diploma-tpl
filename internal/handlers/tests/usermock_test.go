package handlers__test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers/servicemock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Create(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                   string
		inputMetod             string
		inputEndpoint          string
		inputBody              string
		expectedStatusCode     int
		expectedResponseBody   string
		expectedHeader2        string
		expectedHeaderContent2 string
	}{
		// определяем все тесты
		{
			name:                   "Positive test for user registration",
			inputMetod:             "POST",
			inputEndpoint:          "/api/user/register",
			inputBody:              `{ "login": "dimma", "password": "12345" }`,
			expectedStatusCode:     http.StatusOK,
			expectedHeader2:        "Authorization",
			expectedHeaderContent2: "Bearer",
		},
		{
			name:               "Negative test user registration - wrong JSON",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimma, "password": "12345" }`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Negative test user registration - login exist",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimma2login", "password": "12345" }`,
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "Negative test user registration - server error",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimmaServErr", "password": "12345" }`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	s := &servicemock.User{}
	h := handlers.NewUserHandler(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// конфигурирование запроса
			rout := chi.NewRouter()
			rout.Post("/api/user/register", h.Create)
			// запрос
			request := httptest.NewRequest(tCase.inputMetod, tCase.inputEndpoint, bytes.NewBufferString(tCase.inputBody))
			/// создание запроса
			w := httptest.NewRecorder()
			// запуск
			rout.ServeHTTP(w, request)
			// оценка результатов
			assert.Equal(t, tCase.expectedStatusCode, w.Code)
			if w.Code == http.StatusOK {
				assert.Contains(t, w.Header().Get(tCase.expectedHeader2), tCase.expectedHeaderContent2)
			}

		})
	}
}

func TestHandler_CheckAuthorization(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                   string
		inputMetod             string
		inputEndpoint          string
		inputBody              string
		expectedStatusCode     int
		expectedResponseBody   string
		expectedHeader2        string
		expectedHeaderContent2 string
	}{
		// определяем все тесты
		{
			name:                   "Positive test for user registration",
			inputMetod:             "POST",
			inputEndpoint:          "/api/user/register",
			inputBody:              `{ "login": "dimma", "password": "12345" }`,
			expectedStatusCode:     http.StatusOK,
			expectedHeader2:        "Authorization",
			expectedHeaderContent2: "Bearer",
		},
		{
			name:               "Negative test user registration - wrong JSON",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimma, "password": "12345" }`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Negative test user registration - login not exist",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimma3login", "password": "12345" }`,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Negative test user registration - server error",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimmaServErr", "password": "12345" }`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	s := &servicemock.User{}
	h := handlers.NewUserHandler(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// конфигурирование тестового сервера
			rout := chi.NewRouter()
			rout.Post("/api/user/register", h.CheckAuthorization)
			// конфигурирование запроса
			request := httptest.NewRequest(tCase.inputMetod, tCase.inputEndpoint, bytes.NewBufferString(tCase.inputBody))
			// создание запроса
			w := httptest.NewRecorder()
			// запуск
			rout.ServeHTTP(w, request)
			// оценка результатов
			assert.Equal(t, tCase.expectedStatusCode, w.Code)
			if w.Code == http.StatusOK {
				assert.Contains(t, w.Header().Get(tCase.expectedHeader2), tCase.expectedHeaderContent2)
			}

		})
	}
}
