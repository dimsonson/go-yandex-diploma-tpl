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

func TestHandler_C1eate(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputMetod           string
		inputEndpoint        string
		inputBody            string
		expectedStatusCode   int
		expectedResponseBody string
		//expectedHeader1        string
		expectedHeader2 string
		//expectedHeaderContent1 string
		expectedHeaderContent2 string
	}{
		// определяем все тесты
		{
			name:               "OK (positive) test for user registration",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimma", "password": "12345" }`,
			expectedStatusCode: http.StatusOK,
			// expectedHeader1:        "Content-Type",
			// expectedHeaderContent1: "text/plain",
			expectedHeader2:        "Authorization",
			expectedHeaderContent2: "Bearer",
		},
		{
			name:               "Negative test user registration - wrong JSON",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimma, "password": "12345" }`,
			expectedStatusCode: http.StatusBadRequest,
			//expectedHeader1:        "Content-Type",
			//expectedHeaderContent1: "text/plain; charset=utf-8",
		},
		{
			name:               "Negative test user registration - login exist",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimma2login", "password": "12345" }`,
			expectedStatusCode: http.StatusConflict,
			//expectedHeader1:        "Content-Type",
			//expectedHeaderContent1: "text/plain; charset=utf-8",
		},
		{
			name:               "Negative test user registration - server error",
			inputMetod:         "POST",
			inputEndpoint:      "/api/user/register",
			inputBody:          `{ "login": "dimmaServErr", "password": "12345" }`,
			expectedStatusCode: http.StatusInternalServerError,
			//expectedHeader1:        "Content-Type",
			//expectedHeaderContent1: "text/plain; charset=utf-8",
		},
	}
	s := &servicemock.User{}
	h := handlers.NewUserHandler(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// конфигурирование тестового сервера
			rout := chi.NewRouter()
			rout.Post("/api/user/register", h.Create)
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
