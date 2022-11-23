package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/shopspring/decimal"

	"github.com/rs/zerolog/log"
)

// интерфейс методов бизнес логики User
type User interface {
	Create(ctx context.Context, dc models.DecodeLoginPair) (err error)
	CheckAuthorization(ctx context.Context, dc models.DecodeLoginPair) (err error)
}

// интерфейс методов бизнес логики Order
type Order interface {
	Load(ctx context.Context, login string, orderNum string) (err error)
	List(ctx context.Context, login string) (ec []models.OrdersList, err error)
}

// интерфейс методов бизнес логики type Balance
type Balance interface {
	Status(ctx context.Context, login string) (ec models.LoginBalance, err error)
	NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error)
	WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error)
}

// структура для конструктура обработчика User
type UserHandler struct {
	User User
}
// структура для конструктура обработчика Order
type OrderHandler struct {
	Order Order
}

// структура для конструктура обработчика Balance
type BalanceHandler struct {
	Balance Balance
}

// конструктор обработчика User
func NewUserHandler(hUser User) *UserHandler {
	return &UserHandler{
		hUser,
	}
}

// конструктор обработчика Order
func NewOrderHandler(hOrder Order) *OrderHandler {
	return &OrderHandler{
		hOrder,
	}
}

// конструктор обработчика Balance
func NewBalanceHandler(hBalance Balance) *BalanceHandler {
	return &BalanceHandler{
		hBalance,
	}
}

// обработка всех остальных запросов и возврат кода 400
func (handler UserHandler) IncorrectRequests(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "request incorrect", http.StatusBadRequest)
	log.Printf("request incorrect probably no endpoint for path")
}

// регистрация пользователя: HTTPзаголовок Authorization
func (handler UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	// наследуем контекcт запроса r *http.Request, оснащая его Timeout
	ctx, cancel := context.WithTimeout(r.Context(), settings.StorageTimeout)
	// освобождаем ресурс
	defer cancel()
	// десериализация тела запроса
	dc := models.DecodeLoginPair{}
	err := json.NewDecoder(r.Body).Decode(&dc)
	if err != nil {
		log.Printf("Unmarshal error: %s", err)
		http.Error(w, "invalid JSON structure received", http.StatusBadRequest)
		return
	}
	// пишем пару логин:пароль в хранилище
	err = handler.User.Create(ctx, dc)
	// если логин существует возвращаем статус 409, если иная ошибка - 500, если без ошибок -200
	switch {
	case err != nil && strings.Contains(err.Error(), "login exist"):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		_, tokenString, _ := settings.TokenAuth.Encode(map[string]interface{}{"login": dc.Login})
		w.Header().Set("Authorization", "Bearer "+tokenString)
		w.WriteHeader(http.StatusOK)
	}
}

// аутентификация пользователя: HTTPзаголовок Authorization
func (handler UserHandler) CheckAuthorization(w http.ResponseWriter, r *http.Request) {
	// наследуем контекcт запроса r *http.Request, оснащая его Timeout
	ctx, cancel := context.WithTimeout(r.Context(), settings.StorageTimeout)
	// освобождаем ресурс
	defer cancel()
	// десериализация тела запроса
	dc := models.DecodeLoginPair{}
	err := json.NewDecoder(r.Body).Decode(&dc)
	if err != nil {
		log.Printf("unmarshal error HandlerUserAuth: %s", err)
		http.Error(w, "invalid JSON structure received", http.StatusBadRequest)
		return
	}
	// проверяем пару логин:пароль в хранилище
	err = handler.User.CheckAuthorization(ctx, dc)
	// если логин существует и пароль ок возвращаем статус 200, если иная ошибка - 500, если пара неверна - 401
	switch {
	case err != nil && strings.Contains(err.Error(), "login or password not exist"):
		w.WriteHeader(http.StatusUnauthorized)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		_, tokenString, _ := settings.TokenAuth.Encode(map[string]interface{}{"login": dc.Login})
		w.Header().Set("Authorization", "Bearer "+tokenString)
		w.WriteHeader(http.StatusOK)
	}
}

// загрузка пользователем номера заказа для расчёта
func (handler OrderHandler) Load(w http.ResponseWriter, r *http.Request) {
	// наследуем контекcт запроса r *http.Request, оснащая его Timeout
	ctx, cancel := context.WithTimeout(r.Context(), settings.StorageTimeout)
	// освобождаем ресурс
	defer cancel()
	// читаем Body
	bs, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		log.Printf("boby read HandlerNewOrderLoad error :%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b := string(bs)
	// проверяем на алгоритм Луна, если не ок, возвращаем 422
	err = goluhn.Validate(b)
	if err != nil {
		log.Printf("luhn algo check error :%s", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// получаем значение login из контекста запроса
	_, tokenString, _ := jwtauth.FromContext(r.Context())
	// проверяем пару логин:пароль в хранилище
	err = handler.Order.Load(ctx, tokenString["login"].(string), b)
	// если ордер существует от этого пользователя - статус 200, если иная ошибка - 500
	// если от другого пользователя - 409
	// если нет ошибок - 202
	switch {
	case err != nil && strings.Contains(err.Error(), "order number from this login already exist"):
		w.WriteHeader(http.StatusOK)
	case err != nil && strings.Contains(err.Error(), "the same order number was loaded by another customer"):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusAccepted)
	}

}

// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
func (handler OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	// наследуем контекcт запроса r *http.Request, оснащая его Timeout
	ctx, cancel := context.WithTimeout(r.Context(), settings.StorageTimeout)
	// освобождаем ресурс
	defer cancel()
	// получаем значение login из контекста запроса
	_, tokenString, _ := jwtauth.FromContext(r.Context())
	// получаем слайс структур и ошибку
	ec, err := handler.Order.List(ctx, tokenString["login"].(string))
	// устанавливаем заголовок
	w.Header().Set("Content-Type", "application/json")
	// 200 - при ошибке nil, 204 - при ошибке "no records for this login", 500 - при иных ошибках сервиса
	switch {
	case err != nil && strings.Contains(err.Error(), "no orders for this login"):
		w.WriteHeader(http.StatusNoContent)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		//устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)
		// сериализуем и пишем тело ответа
		json.NewEncoder(w).Encode(ec)

	}
}

// получение текущего баланса счёта баллов лояльности пользователя
func (handler BalanceHandler) Status(w http.ResponseWriter, r *http.Request) {
	// наследуем контекcт запроса r *http.Request, оснащая его Timeout
	ctx, cancel := context.WithTimeout(r.Context(), settings.StorageTimeout)
	// освобождаем ресурс
	defer cancel()
	// получаем значение login из контекста запроса
	_, tokenString, _ := jwtauth.FromContext(r.Context())
	// получаем слайс структур и ошибку
	ec, err := handler.Balance.Status(ctx, tokenString["login"].(string))
	// устанавливаем заголовок
	w.Header().Set("Content-Type", "application/json")
	// 200 - при ошибке nil, 500 - при иных ошибках сервиса
	switch {
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		//устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)
		// сериализуем и пишем тело ответа
		json.NewEncoder(w).Encode(ec)
	}
}

// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
func (handler BalanceHandler) NewWithdrawal(w http.ResponseWriter, r *http.Request) {
	// наследуем контекcт запроса r *http.Request, оснащая его Timeout
	ctx, cancel := context.WithTimeout(r.Context(), settings.StorageTimeout)
	// освобождаем ресурс
	defer cancel()
	// десериализация тела запроса
	dc := models.NewWithdrawal{}
	err := json.NewDecoder(r.Body).Decode(&dc)
	if err != nil {
		log.Printf("unmarshal error HandlerNewWithdrawal: %s", err)
		http.Error(w, "invalid JSON structure received", http.StatusBadRequest)
		return
	}
	// проверяем на алгоритм Луна, если не ок, возвращаем 422
	err = goluhn.Validate(dc.Order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// получаем значение login из контекста запроса
	_, tokenString, _ := jwtauth.FromContext(r.Context())
	// отпправляем на списание
	err = handler.Balance.NewWithdrawal(ctx, tokenString["login"].(string), dc)
	// 200 - при ошибке nil, 500 - при иных ошибках сервиса, 422 - проверка Луна не ок
	// 402 - если получена ошибка "insufficient funds"
	switch {
	case err != nil && strings.Contains(err.Error(), "insufficient funds"):
		w.WriteHeader(http.StatusPaymentRequired)
	case err != nil && strings.Contains(err.Error(), "new order number already exist"):
		w.WriteHeader(http.StatusUnprocessableEntity)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// получение информации о выводе средств с накопительного счёта пользователем
func (handler BalanceHandler) WithdrawalsList(w http.ResponseWriter, r *http.Request) {
	// наследуем контекcт запроса r *http.Request, оснащая его Timeout
	ctx, cancel := context.WithTimeout(r.Context(), settings.StorageTimeout)
	// освобождаем ресурс
	defer cancel()
	// получаем значение login из контекста запроса
	_, tokenString, _ := jwtauth.FromContext(r.Context())
	// получаем слайс структур и ошибку
	ec, err := handler.Balance.WithdrawalsList(ctx, tokenString["login"].(string))
	// устанавливаем заголовок
	w.Header().Set("Content-Type", "application/json")
	// 200 - при ошибке nil, кодирование, 500 - при иных ошибках сервиса, 204 - если получена ошибка "no records"
	switch {
	case err != nil && strings.Contains(err.Error(), "no records"):
		w.WriteHeader(http.StatusNoContent)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		// устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)
		// сериализуем и пишем тело ответа
		json.NewEncoder(w).Encode(ec)
	}
}
