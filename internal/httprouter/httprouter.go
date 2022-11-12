package httprouter

import (
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// маршрутизатор запросов
func NewRouter(hn *handlers.Handler) chi.Router {
	// chi роутер
	rout := chi.NewRouter()
	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	rout.Use(middleware.Logger)
	rout.Use(middleware.Recoverer)
	// дополнительный middleware
	rout.Use(middlewareGzip)
	// rout.Use(middlewareCookie)

	// маршрут DELETE "/api/user/urls" пакетное удаление коротки ссылок
	//rout.Delete("/api/user/urls", hn.HandlerDeleteBatch)
	
	// маршрут POST "/api/shorten/batch" пакетная выдача коротких ссылок
	//rout.Post("/api/shorten/batch", hn.HandlerCreateBatchJSON)
	
	// маршрут GET "/ping" проверка доступности PostgreSQL
	//rout.Get("/ping", hn.HandlerSQLping)
	
	// маршрут GET "/api/user/urls" id в URL
	//rout.Get("/api/user/urls", hn.HandlerGetUserURLs)
	
	// маршрут GET "/{id}" id в URL
	//rout.Get("/{id}", hn.HandlerGetShortURL)
	
	// маршрут POST "/api/shorten" c JSON в теле запроса
	//rout.Post("/api/shorten", hn.HandlerCreateShortJSON)
	
	// маршрут POST "/" с текстовым URL в теле запроса
	//rout.Post("/", hn.HandlerCreateShortURL)
	
	// возврат ошибки 400 для всех остальных запросов
	//rout.HandleFunc("/*", hn.IncorrectRequests)

	return rout
}
