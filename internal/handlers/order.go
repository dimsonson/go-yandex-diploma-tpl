package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/shopspring/decimal"

	"github.com/rs/zerolog/log"
)

// интерфейс методов бизнес логики Order
type OrderServiceProvider interface {
	Load(ctx context.Context, login string, orderNum string) (err error)
	List(ctx context.Context, login string) (ec []models.OrdersList, err error)
}

// структура для конструктура обработчика Order
type OrderHandler struct {
	service OrderServiceProvider
}

// конструктор обработчика Order
func NewOrderHandler(hOrder OrderServiceProvider) *OrderHandler {
	return &OrderHandler{
		hOrder,
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
		log.Printf("body read HandlerLoad error :%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b := string(bs)
	// проверяем, пришли ли цифры в номере заказа
	_, err = strconv.Atoi(b)
	if err != nil {
		log.Printf("digits check HandlerLoad error :%s", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// проверяем на алгоритм Луна, если не ок, возвращаем 422
	err = goluhn.Validate(b)
	if err != nil {
		log.Printf("luhn algo check HandlerLoad error :%s", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// получаем значение login из контекста запроса
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Printf("FromContext error HandlerLoad: %s", err)
		http.Error(w, "order handling error", http.StatusInternalServerError)
		return
	}
	// получаем значение из интерфейса
	login, ok := claims["login"].(string)
	if !ok {
		log.Printf("interface assertion error HandlerLoad: %s", err)
		http.Error(w, "order handling error", http.StatusInternalServerError)
		return
	}
	// проверяем пару логин:пароль в хранилище
	err = handler.service.Load(ctx, login, b)
	// если ордер существует от этого пользователя - статус 200, если иная ошибка - 500
	// если от другого пользователя - 409 // если нет ошибок - 202
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
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Printf("FromContext error HandlerList: %s", err)
		http.Error(w, "order handling error", http.StatusInternalServerError)
		return
	}
	// получаем значение из интерфейса
	login, ok := claims["login"].(string)
	if !ok {
		log.Printf("interface assertion error HandlerList: %s", err)
		http.Error(w, "order handling error", http.StatusInternalServerError)
		return
	}
	// устанавливаем заголовок
	w.Header().Set("Content-Type", "application/json")
	// получаем слайс структур и ошибку
	ec, err := handler.service.List(ctx, login)
	// 200 - при ошибке nil, 204 - при ошибке "no records for this login", 500 - при иных ошибках сервиса
	switch {
	case err != nil && strings.Contains(err.Error(), "no orders for this login"):
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
