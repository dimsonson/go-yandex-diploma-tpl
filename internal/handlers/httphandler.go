package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/go-chi/jwtauth"
)

// интерфейс методов бизнес логики
type Services interface {
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

// регистрация пользователя; HTTPзаголовок Authorization.
func (hn Handler) HandlerNewUserReg(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "server error", http.StatusInternalServerError)
}

// аутентификация пользователя
func (hn Handler) HandlerUserAuth(w http.ResponseWriter, r *http.Request) {
	// десериализация тела запроса
	dc := DecodeLoginPair{}
	err := json.NewDecoder(r.Body).Decode(&dc)
	if err != nil {
		log.Printf("Unmarshal error: %s", err)
		http.Error(w, "invalid JSON structure received", http.StatusBadRequest)
	}

	_, tokenString, _ := settings.TokenAuth.Encode(map[string]interface{}{"login": dc.Login})
	fmt.Println("tokenString", tokenString)

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %v", tokenString))
	fmt.Println("w.Header()", w.Header())

	w.WriteHeader(http.StatusOK)
	
}

// загрузка пользователем номера заказа для расчёта
func (hn Handler) HandlerNewOrderLoad(w http.ResponseWriter, r *http.Request) {
	// получаем значение iserid из контекста запроса
	token, tokenString, err := jwtauth.FromContext(r.Context())
	fmt.Println("token", token)
	fmt.Println("tokenString", tokenString)
	fmt.Println("err", err)

	http.Error(w, "server error", http.StatusInternalServerError)
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

type DecodeLoginPair struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
