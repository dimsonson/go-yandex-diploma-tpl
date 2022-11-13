package httprouter

import (
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
)

// маршрутизатор запросов
func NewRouter(hn *handlers.Handler) chi.Router {
	// chi роутер
	rout := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	rout.Use(middleware.Logger)
	rout.Use(middleware.Recoverer)
	// дополнительный middleware gzip
	rout.Use(middlewareGzip)

	// Protected routes
	rout.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtauth.Verifier(settings.TokenAuth))
		// Handle valid / invalid tokens
		r.Use(jwtauth.Authenticator)
		// загрузка пользователем номера заказа для расчёта
		r.Post("/api/user/orders", hn.HandlerNewOrderLoad)
		// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
		r.Get("/api/user/orders", hn.HandlerGetOrdersList)
		// получение текущего баланса счёта баллов лояльности пользователя
		r.Get("/api/user/balance", hn.HandlerGetUserBalance)
		// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
		r.Post("/api/user/balance/withdraw", hn.HandlerNewWithdrawal)
		// получение информации о выводе средств с накопительного счёта пользователем
		r.Get("/api/user/balance/withdrawals", hn.HandlerGetWithdrawalsList)

	})

	// Public routes
	rout.Group(func(r chi.Router) {
		// регистрация пользователя: HTTPзаголовок Authorization
		r.Post("/api/user/register", hn.HandlerNewUserReg)
		// аутентификация пользователя: HTTPзаголовок Authorization
		r.Post("/api/user/login", hn.HandlerUserAuth)
	})

	// возврат ошибки 400 для всех остальных запросов
	rout.HandleFunc("/*", hn.IncorrectRequests)

	return rout
}
