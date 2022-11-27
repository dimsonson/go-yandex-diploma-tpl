package handlers__test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers/servicemock"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Load(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name               string
		inputMetod         string
		inputEndpoint      string
		inputBody          string
		expectedStatusCode int
		inputLogin         string
	}{
		// определяем все тесты
		{
			name:               "Positive test for order load",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma",
			inputBody:          "1235489802",
			expectedStatusCode: http.StatusAccepted,
		},
		{
			name:               "Negative test order load - not every sign from order number is digit",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma",
			inputBody:          "1235AAAA489802",
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:               "Negative test order load - luhn algo check isn't ok",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma",
			inputBody:          "123548980260",
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:               "Negative test order load - order alredy exist for this login",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma2login",
			inputBody:          "1235489802",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Negative test order load - order alredy exist for another login",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma3",
			inputBody:          "1235489802",
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "Negative test order load - any other internal error",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma8",
			inputBody:          "1235489802",
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	s := &servicemock.Order{}
	h := handlers.NewOrderHandler(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// конфигурирование тестового сервера
			rout := chi.NewRouter()
			rout.Post("/api/user/orders", h.Load)
			// конфигурирование запроса
			request := httptest.NewRequest(tCase.inputMetod, tCase.inputEndpoint, bytes.NewBufferString(tCase.inputBody))
			// контекст логина
			tkn := jwt.New()
			tkn.Set(`login`, tCase.inputLogin)
			rctx := jwtauth.NewContext(request.Context(), tkn, nil)
			request = request.WithContext(rctx)
			// создание запроса
			w := httptest.NewRecorder()
			w.Header().Set("Authorization", "Bearer "+tCase.inputLogin)
			// запуск
			rout.ServeHTTP(w, request)
			// оценка результатов
			assert.Equal(t, tCase.expectedStatusCode, w.Code)
		})
	}
}

func TestHandler_List(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name          string
		inputMetod    string
		inputEndpoint string

		expectedStatusCode int
		inputLogin         string
	}{
		// определяем все тесты
		{
			name:               "Positive test for order list report",
			inputMetod:         http.MethodGet,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Negative test order list - no orders for report",
			inputMetod:         http.MethodGet,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma2login",
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "Negative test order list - any other internal error",
			inputMetod:         http.MethodGet,
			inputEndpoint:      "/api/user/orders",
			inputLogin:         "dimma8",
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	s := &servicemock.Order{}
	h := handlers.NewOrderHandler(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// конфигурирование тестового сервера
			rout := chi.NewRouter()
			rout.Get("/api/user/orders", h.List)
			// конфигурирование запроса
			request := httptest.NewRequest(tCase.inputMetod, tCase.inputEndpoint, nil)
			// контекст логина
			tkn := jwt.New()
			tkn.Set(`login`, tCase.inputLogin)
			rctx := jwtauth.NewContext(request.Context(), tkn, nil)
			request = request.WithContext(rctx)
			// создание запроса
			w := httptest.NewRecorder()
			w.Header().Set("Authorization", "Bearer "+tCase.inputLogin)
			// запуск
			rout.ServeHTTP(w, request)
			// оценка результатов
			assert.Equal(t, tCase.expectedStatusCode, w.Code)
		})
	}
}
