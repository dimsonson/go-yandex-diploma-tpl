package handlers

import (
	"bytes"
	"net/http/httptest"
	"testing"

	mock_service "github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers/mocks"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"

	//"github.com/dimsonson/go-yandex-diploma-tpl/internal/services"
	//mock_storage "github.com/dimsonson/go-yandex-diploma-tpl/internal/services/mocks"
	"github.com/go-chi/chi/v5"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -destination=mock_service_test.go -package=handlers_test . Services

func TestHandler_HandlerNewUserReg(t *testing.T) {
	type mockBehavior func(s *mock_service.MockServices, dc models.DecodeLoginPair)
	//type mockBehavior func(s *mock_storage.MockStorageProvider, login string, passwH string) //(err error)

	// определяем структуру теста
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                 string
		inputMetod           string
		inputEndpoint        string
		inputBody            string
		dc                   models.DecodeLoginPair
		mockBehavior         mockBehavior
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
			dc: models.DecodeLoginPair{Login: "dimma", Password: "12345"},
			mockBehavior: func(s *mock_service.MockServices, dc models.DecodeLoginPair) {
				s.EXPECT().ServiceCreateNewUser(nil, dc).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "Authorization",
			expectedContentType:  "text/plain; charset=utf-8",
		},
	}

	for _, tCase := range tests {
		// запускаем каждый тест
		t.Run(tCase.name, func(t *testing.T) {
			// Init gomock controller
			c := gomock.NewController(t)
			defer c.Finish()
			// Init Dependencies
			srvc := mock_service.NewMockServices(c)
			//stor := mock_storage.NewMockStorageProvider(c)
			tCase.mockBehavior(srvc, tCase.dc)
			//service := &services.Services{Storage: stor, CalcSys: "http://localhost:8080"}
			handler := NewHandler(srvc)
			// Test Server
			rout := chi.NewRouter()
			rout.Post("/api/user/register", handler.HandlerNewUserReg)
			// create Test Request
			w := httptest.NewRecorder()
			request := httptest.NewRequest(tCase.inputMetod, tCase.inputEndpoint, bytes.NewBufferString(tCase.inputBody))
			// perform Test Request
			rout.ServeHTTP(w, request)
			// Assert results
			assert.Equal(t, tCase.expectedStatusCode, w.Code)
			assert.Contains(t, tCase.expectedResponseBody, w.Body.String())
		})
	}
}
