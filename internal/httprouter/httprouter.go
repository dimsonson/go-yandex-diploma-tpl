package httprouter

import (
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

// маршрутизатор запросов
func NewRouter(userHandler *handlers.UserHandler, orderHandler *handlers.OrderHandler, balanceHandler *handlers.BalanceHandler) chi.Router {
	// chi роутер
	rout := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	rout.Use(middleware.Logger)
	rout.Use(middleware.Recoverer)
	// дополнительный middleware gzip
	rout.Use(middlewareGzip)

	// защищенные пути
	rout.Group(func(r chi.Router) {
		// поиск, верифицирование, валидация JWT токенов
		r.Use(jwtauth.Verifier(settings.TokenAuth))
		// обрабочик валидный / не валидный токен
		r.Use(jwtauth.Authenticator)
		// загрузка пользователем номера заказа для расчёта
		r.Post("/api/user/orders", orderHandler.Load)
		// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
		r.Get("/api/user/orders", orderHandler.List)
		// получение текущего баланса счёта баллов лояльности пользователя
		r.Get("/api/user/balance", balanceHandler.Status)
		// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
		r.Post("/api/user/balance/withdraw", balanceHandler.NewWithdrawal)
		// получение информации о выводе средств с накопительного счёта пользователем
		r.Get("/api/user/withdrawals", balanceHandler.WithdrawalsList)

	})

	// публичные пути
	rout.Group(func(r chi.Router) {
		// регистрация пользователя: HTTPзаголовок Authorization
		r.Post("/api/user/register", userHandler.Create)
		// аутентификация пользователя: HTTPзаголовок Authorization
		r.Post("/api/user/login", userHandler.CheckAuthorization)
	})

	// возврат ошибки 400 для всех остальных запросов
	rout.HandleFunc("/*", userHandler.IncorrectRequests)

	return rout
}
