package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/go-chi/jwtauth"
)

// интерфейс методов бизнес логики
type Services interface {
	ServiceCreateNewUser(settings.DecodeLoginPair) (err error)
	ServiceAuthorizationCheck(dc settings.DecodeLoginPair) (err error)
	ServiceNewOrderLoad(login string, order_num string) (err error)
}

// структура для конструктура обработчика
type Handler struct {
	service Services
}

// конструктор обработчика
func NewHandler(s Services) *Handler {
	return &Handler{
		s,
	}
}

// обработка всех остальных запросов и возврат кода 400
func (hn Handler) IncorrectRequests(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "request incorrect", http.StatusBadRequest)
}

// регистрация пользователя: HTTPзаголовок Authorization
func (hn Handler) HandlerNewUserReg(w http.ResponseWriter, r *http.Request) {
	// десериализация тела запроса
	dc := settings.DecodeLoginPair{}
	err := json.NewDecoder(r.Body).Decode(&dc)
	if err != nil {
		log.Printf("Unmarshal error: %s", err)
		http.Error(w, "invalid JSON structure received", http.StatusBadRequest)
		return
	}
	// пишем пару логин:пароль в хранилище
	err = hn.service.ServiceCreateNewUser(dc)
	// если логин существует возвращаем статус 409, если иная ошибка - 500, если без ошибок -200
	switch {
	case err != nil && strings.Contains(err.Error(), "login exist"):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// аутентификация пользователя: HTTPзаголовок Authorization
func (hn Handler) HandlerUserAuth(w http.ResponseWriter, r *http.Request) {
	// десериализация тела запроса
	dc := settings.DecodeLoginPair{}
	err := json.NewDecoder(r.Body).Decode(&dc)
	if err != nil {
		log.Printf("Unmarshal error: %s", err)
		http.Error(w, "invalid JSON structure received", http.StatusBadRequest)
		return
	}
	// проверяем пару логин:пароль в хранилище
	err = hn.service.ServiceAuthorizationCheck(dc)
	// если логин существует и пароль ок возвращаем статус 200, если иная ошибка - 500, если пара неверна - 401
	switch {
	case err != nil && strings.Contains(err.Error(), "login or password incorrect"):
		w.WriteHeader(http.StatusUnauthorized)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		_, tokenString, _ := settings.TokenAuth.Encode(map[string]interface{}{"login": dc.Login})
		fmt.Println("tokenString", tokenString)
		w.Header().Set("Authorization", fmt.Sprintf("Bearer %v", tokenString))
		fmt.Println("w.Header()", w.Header())
		w.WriteHeader(http.StatusOK)
	}
}

// загрузка пользователем номера заказа для расчёта
func (hn Handler) HandlerNewOrderLoad(w http.ResponseWriter, r *http.Request) {
	// читаем Body
	bs, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b := string(bs)
	// проверяем на алгоритм Луна, если не ок, возвращаем 422
	err = goluhn.Validate(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// получаем значение login из контекста запроса
	_, tokenString, _ := jwtauth.FromContext(r.Context())
	// проверяем пару логин:пароль в хранилище
	err = hn.service.ServiceNewOrderLoad(tokenString["login"].(string), b)
	// если ордер существует от этого пользователя - статус 200, если иная ошибка - 500
	// если от другого пользователя - 409
	// если нет ошибок - 202
	switch {
	case err != nil && strings.Contains(err.Error(), "customer order already exist"):
		w.WriteHeader(http.StatusOK)
	case err != nil && strings.Contains(err.Error(), "same order number was loaded by another customer"):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusAccepted)
	}

}

// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
func (hn Handler) HandlerGetOrdersList(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "server error", http.StatusInternalServerError)
}

// получение текущего баланса счёта баллов лояльности пользователя
func (hn Handler) HandlerGetUserBalance(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "server error", http.StatusInternalServerError)
}

// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
func (hn Handler) HandlerNewUserAccWithdrawal(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "server error", http.StatusInternalServerError)
}

// получение информации о выводе средств с накопительного счёта пользователем
func (hn Handler) HandlerGetWithdrawalsList(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "server error", http.StatusInternalServerError)
}
