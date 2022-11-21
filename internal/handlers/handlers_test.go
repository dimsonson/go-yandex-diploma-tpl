package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers"
	//"github.com/dimsonson/go-yandex-diploma-tpl/internal/httprouter"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)




func TestHandlerCreateShortURL(t *testing.T) {
	// определяем структуру теста
	type request struct {
		metod    string
		endpoint string
		body     string
	}

	type want struct {
		code         int
		responseBody string
		contentType  string
	}

	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		req  request
		want want
	}{
		// определяем все тесты
		{
			name: "positive test POST",
			req: request{
				metod:    "POST",
				endpoint: "/",
				body:     "https://pkg.go.dev/io#Reader",
			},
			want: want{
				code:         201,
				responseBody: "/0",
				contentType:  "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {

			//создаем тестирующий запрос
			request := httptest.NewRequest(tt.req.metod, tt.req.endpoint, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
            h := http.HandlerFunc(handlers.NewHandler()

			/* //создаем тестирующий запрос
			req := httptest.NewRequest(tt.req.metod, "/", strings.NewReader("http://localhost:8080/"))

			// создаём новый Recorder

			w := httptest.NewRecorder()

			// определяем хендлер
			s := storage.NewMapStorage(make(map[string]string), make(map[string]string), make(map[string]bool))
			srvs := services.NewService(s, "http://localhost:8080/")
			h := handlers.NewHandler(srvs, "http://localhost:8080/")
			r := httprouters.NewRouter(h)
			//h := http.HandlerFunc(handlers.)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", strings.TrimPrefix(tt.req.endpoint, "/"))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), settings.CtxKeyUserID, "5e7cb52e-691d-4f46-bc1c-7ae1616a59ff"))

			//h := http.HandlerFunc(handlers.NewHandler())

			// запускаем сервер
			r.ServeHTTP(w, req)
			resp := w.Result()

			// проверяем код ответа
			if resp.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// получаем и проверяем тело ответа
			defer resp.Body.Close()
			resBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			// проверка содержания строки в теле ответа
			assert.Containsf(t, string(resBody), tt.want.response, "error message %s", "formatted")

			// заголовок ответа
			if resp.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, resp.Header.Get("Content-Type"))
			} */
		})
	}
}

func TestHandlerCreateShortJSON(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}

	type req struct {
		metod    string
		endpoint string
		body     string
	}

	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		req  req
		want want
	}{
		// определяем все тесты
		{
			name: "POST - URL Successfully created",
			req: req{
				metod:    "POST",
				endpoint: "/api/shorten",
				body:     `{"url":"https://yandex.ru/search/?text=AToi+go&lr=213"}`,
			},
			want: want{
				code:        201,
				response:    "/0",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "POST empty body",
			req: req{
				metod:    "POST",
				endpoint: "/api/shorten",
				body:     "",
			},
			want: want{
				code:        400,
				response:    "invalid URL received",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "POST wrong url",
			req: req{
				metod:    "POST",
				endpoint: "/api/shorten",
				body:     `{"url":"htpandexru/search/?text=AToi+go&lr=213"}`,
			},
			want: want{
				code:        400,
				response:    "invalid URL received",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {

			//создаем тестирующий запрос
			req := httptest.NewRequest(tt.req.metod, tt.req.endpoint, strings.NewReader(tt.req.body))
			req = req.WithContext(context.WithValue(req.Context(), settings.CtxKeyUserID, "5e7cb52e-691d-4f46-bc1c-7ae1616a59ff"))
			// создаём новый Recorder

			w := httptest.NewRecorder()

			// определяем хендлер
			s := storage.NewMapStorage(make(map[string]string), make(map[string]string), make(map[string]bool))
			srvs := services.NewService(s, "http://localhost:8080/")
			h := handlers.NewHandler(srvs, "http://localhost:8080/")
			//r := httprouters.NewRouter(h)
			h.HandlerCreateShortJSON(w, req)

			// запускаем сервер
			//r.ServeHTTP(w, req)
			resp := w.Result()

			// проверяем код ответа
			if resp.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// получаем и проверяем тело ответа
			defer resp.Body.Close()
			resBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			// проверка содержания строки в теле ответа
			assert.Containsf(t, string(resBody), tt.want.response, "error message %s", "formatted")

			// заголовок ответа
			if resp.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestHandlerGetUserURLs(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}

	type req struct {
		metod    string
		endpoint string
		cookie   http.Cookie
		id       string
		body     string
	}

	//var cook http.Cookie
	cook := http.Cookie{
		Name:   "token",
		Value:  "35653763623532652d363931642d346634362d626331632d3761653136313661353966663fe75fb6b45bd519a5e87f62c5507aff32f4410bed855e9c65628b7b9eee35b6",
		MaxAge: 300,
	}
	cookBad := http.Cookie{}

	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		req  req
		want want
	}{
		// определяем все тесты
		{
			name: "GET - set of URLs Successfully created",
			req: req{
				metod:    "GET",
				endpoint: "/api/user/urls",
				cookie:   cook,
				id:       "5e7cb52e-691d-4f46-bc1c-7ae1616a59ff",
				body:     `{"url":"https://yandex.ru/search/?text=AToi+go&lr=213"}`,
			},
			want: want{
				code:        200,
				response:    "[",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "GET wrong cookie",
			req: req{
				metod:    "GET",
				endpoint: "/api/user/urls",
				cookie:   cookBad,
				id:       "5",
				body:     "",
			},
			want: want{
				code:     204,
				response: "",

				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {

			//создаем тестирующий запрос
			req := httptest.NewRequest(tt.req.metod, tt.req.endpoint, strings.NewReader(tt.req.body))
			req = req.WithContext(context.WithValue(req.Context(), settings.CtxKeyUserID, tt.req.id))
			// установим куку в ответ
			req.AddCookie(&cook)

			defer req.Body.Close()

			// создаём новый Recorder

			w := httptest.NewRecorder()

			//ctx := context.Background()
			//ctx, cancel := context.WithTimeout(ctx, settings.StorageTimeout)
			//defer cancel()
			ctx := context.WithValue(req.Context(), settings.CtxKeyUserID, "5e7cb52e-691d-4f46-bc1c-7ae1616a59ff") //tt.req.id) //"5e7cb52e-691d-4f46-bc1c-7ae1616a59ff")
			// определяем хендлер
			s := storage.NewMapStorage(make(map[string]string), make(map[string]string), make(map[string]bool))
			s.StoragePut(ctx, "xyz", "https://pkg.go.dev/github.com/stretchr/testify@v1.8.0/assert#Containsf", "5e7cb52e-691d-4f46-bc1c-7ae1616a59ff")
			srvs := services.NewService(s, "http://localhost:8080/")
			h := handlers.NewHandler(srvs, "http://localhost:8080/")
			//r := httprouters.NewRouter(h)
			h.HandlerGetUserURLs(w, req)
			fmt.Println(s.UserID)
			// запускаем сервер
			//r.ServeHTTP(w, req)
			resp := w.Result()

			// проверяем код ответа
			if resp.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// получаем и проверяем тело ответа
			defer resp.Body.Close()
			resBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			// проверка содержания строки в теле ответа
			assert.Containsf(t, string(resBody), tt.want.response, "error message %s", "formatted")

			// заголовок ответа
			if resp.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestIncorrectReques1ts(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}

	type req struct {
		metod    string
		endpoint string
		body     string
	}

	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		req  req
		want want
	}{
		// определяем все тесты

		{
			name: "GET #2",
			req: req{
				metod:    "PATCH",
				endpoint: "/",
				//body:     "",
			},
			want: want{
				code: 400,
				//response: "",
				//contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {

			//создаем тестирующий запрос
			req := httptest.NewRequest(tt.req.metod, tt.req.endpoint, nil) //strings.NewReader("http://localhost:8080/"))

			// создаём новый Recorder

			w := httptest.NewRecorder()

			// определяем хендлер
			s := storage.NewMapStorage(make(map[string]string), make(map[string]string), make(map[string]bool))
			srvs := services.NewService(s, "http://localhost:8080/")
			h := handlers.NewHandler(srvs, "http://localhost:8080/")
			r := httprouters.NewRouter(h)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", strings.TrimPrefix(tt.req.endpoint, "/"))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			//	h := http.HandlerFunc(handlers.NewHandler())

			// запускаем сервер
			r.ServeHTTP(w, req)
			resp := w.Result()

			defer resp.Body.Close()

			// проверяем код ответа
			if resp.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

		})
	}
}

func TestHandlerGetShortURL(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}

	type req struct {
		metod    string
		endpoint string
		body     string
	}

	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		req  req
		want want
	}{
		// определяем все тесты
		{
			name: "GET #1",
			req: req{
				metod:    "GET",
				endpoint: "/xyz",
				body:     "",
			},
			want: want{
				code:        307,
				response:    "https://pkg.go.dev/github.com/stretchr/testify@v1.8.0/assert#Containsf",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "GET #2",
			req: req{
				metod:    "GET",
				endpoint: "/",
				body:     "",
			},
			want: want{
				code:        400,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(tt.req.metod, tt.req.endpoint, strings.NewReader("http://localhost:8080/"))
			req = req.WithContext(context.WithValue(req.Context(), settings.CtxKeyUserID, "5e7cb52e-691d-4f46-bc1c-7ae1616a59ff"))

			// создаём новый Recorder

			w := httptest.NewRecorder()

			/* 	ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, settings.StorageTimeout)
			defer cancel()
			*/

			// определяем хендлер
			s := storage.NewMapStorage(make(map[string]string), make(map[string]string), make(map[string]bool))
			srvs := services.NewService(s, "http://localhost:8080/")
			h := handlers.NewHandler(srvs, "http://localhost:8080/")
			r := httprouters.NewRouter(h)
			//	h := http.HandlerFunc(handlers.NewHandler())

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", strings.TrimPrefix(tt.req.endpoint, "/"))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			s.StoragePut(req.Context(), "xyz", "https://pkg.go.dev/github.com/stretchr/testify@v1.8.0/assert#Containsf", "5e7cb52e-691d-4f46-bc1c-7ae1616a59ff")
			// запускаем сервер
			r.ServeHTTP(w, req)
			resp := w.Result()

			defer resp.Body.Close()

			// проверяем код ответа
			if resp.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// заголовок ответа
			if resp.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}
