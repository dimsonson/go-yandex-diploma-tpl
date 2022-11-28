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

func TestHandler_Status(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputMetod           string
		inputEndpoint        string
		inputLogin           string
		expectedStatusCode   int
		expectedResponseBody string
		expectedHeader1      string

		expectedHeaderContent1 string
	}{
		// определяем все тесты
		{
			name:                   "Positive test for user balance status",
			inputMetod:             http.MethodGet,
			inputEndpoint:          "/api/user/balance",
			inputLogin:             "dimma",
			expectedStatusCode:     http.StatusOK,
			expectedHeader1:        "Content-Type",
			expectedHeaderContent1: "application/json",
		},
		{
			name:                   "Negative test for user balance status - Server error",
			inputMetod:             http.MethodGet,
			inputEndpoint:          "/api/user/balance",
			inputLogin:             "dimma2",
			expectedStatusCode:     http.StatusInternalServerError,
			expectedHeader1:        "Content-Type",
			expectedHeaderContent1: "application/json",
		},
	}

	s := &servicemock.Balance{}
	h := handlers.NewBalanceHandler(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// конфигурирование тестового сервера
			rout := chi.NewRouter()
			rout.Get(tCase.inputEndpoint, h.Status)
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
			assert.Equal(t, tCase.expectedHeaderContent1, w.Header().Get(tCase.expectedHeader1))

		})
	}
}

func TestHandler_NewWithdrawal(t *testing.T) {
	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                   string
		inputMetod             string
		inputEndpoint          string
		inputLogin             string
		inputBody              string
		expectedStatusCode     int
		expectedResponseBody   string
		expectedHeader1        string
		expectedHeaderContent1 string
	}{
		// определяем все тесты
		{
			name:               "Positive test for user new withdrawal",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/balance/withdraw",
			inputLogin:         "dimma",
			inputBody:          `{"order": "2377225624", "sum":751}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Negative test for user new withdrawal - insufficient funds",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/balance/withdraw",
			inputLogin:         "dimma",
			inputBody:          `{"order": "2377225624", "sum":1751}`,
			expectedStatusCode: http.StatusPaymentRequired,
		},
		{
			name:               "Negative test for user new withdrawal - new order number already exist",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/balance/withdraw",
			inputLogin:         "dimma",
			inputBody:          `{"order": "24564564536456", "sum":751}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:               "Negative test for user new withdrawal - InternalServerError",
			inputMetod:         http.MethodPost,
			inputEndpoint:      "/api/user/balance/withdraw",
			inputLogin:         "dimma2",
			inputBody:          `{"order": "2377225624", "sum":751}`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	s := &servicemock.Balance{}
	h := handlers.NewBalanceHandler(s)

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// конфигурирование тестового сервера
			rout := chi.NewRouter()
			rout.Post(tCase.inputEndpoint, h.NewWithdrawal)
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
