package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	_ "github.com/shopspring/decimal"

	"github.com/rs/zerolog/log"
)

// интерфейс методов бизнес логики User
type User interface {
	Create(ctx context.Context, dc models.DecodeLoginPair) (err error)
	CheckAuthorization(ctx context.Context, dc models.DecodeLoginPair) (err error)
}

// структура для конструктура обработчика User
type UserHandler struct {
	service User
}

// конструктор обработчика User
func NewUserHandler(hUser User) *UserHandler {
	return &UserHandler{
		hUser,
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
	err = handler.service.Create(ctx, dc)
	// если логин существует возвращаем статус 409, если иная ошибка - 500, если без ошибок -200
	switch {
	case err != nil && strings.Contains(err.Error(), "login exist"):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		_, tokenString, err := settings.TokenAuth.Encode(map[string]interface{}{"login": dc.Login})
		if err != nil {
			log.Printf("tokenAuth.Encode error HandlerCreate: %s", err)
			http.Error(w, "login handling error", http.StatusInternalServerError)
			return
		}
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
		log.Printf("unmarshal error HandlerCheckAuthorization: %s", err)
		http.Error(w, "invalid JSON structure received", http.StatusBadRequest)
		return
	}
	// проверяем пару логин:пароль в хранилище
	err = handler.service.CheckAuthorization(ctx, dc)
	// если логин существует и пароль ок возвращаем статус 200, если иная ошибка - 500, если пара неверна - 401
	switch {
	case err != nil && strings.Contains(err.Error(), "login or password not exist"):
		w.WriteHeader(http.StatusUnauthorized)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		_, tokenString, err := settings.TokenAuth.Encode(map[string]interface{}{"login": dc.Login})
		if err != nil {
			log.Printf("tokenAuth.Encode error HandlerCheckAuthorization: %s", err)
			http.Error(w, "login handling error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", "Bearer "+tokenString)
		w.WriteHeader(http.StatusOK)
	}
}
