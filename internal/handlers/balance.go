package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/shopspring/decimal"

	"github.com/rs/zerolog/log"
)

// интерфейс методов бизнес логики type Balance
type Balance interface {
	Status(ctx context.Context, login string) (ec models.LoginBalance, err error)
	NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error)
	WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error)
}

// структура для конструктура обработчика Balance
type BalanceHandler struct {
	Balance Balance
}

// конструктор обработчика Balance
func NewBalanceHandler(hBalance Balance) *BalanceHandler {
	return &BalanceHandler{
		hBalance,
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
