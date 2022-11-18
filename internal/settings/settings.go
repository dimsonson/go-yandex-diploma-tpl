package settings

import (
	"time"

	"github.com/go-chi/jwtauth"
)

// ключ подписи
const SignKey string = "9e9e0b4e6de418b2f84fca35165571c5"

// timeout контекста
const StorageTimeout = 600 * time.Second

// константы цветового вывода в консоль
const (
	ColorBlack  = "\u001b[30m"
	ColorRed    = "\u001b[31m"
	ColorGreen  = "\u001b[32m"
	ColorYellow = "\u001b[33m"
	ColorBlue   = "\u001b[34m"
	ColorReset  = "\u001b[0m"
)

// переменная ключа токена
var TokenAuth *jwtauth.JWTAuth = jwtauth.New("HS256", []byte(SignKey), nil)

// начальный таймаут для горутины запросов к сервису расчета баллов
var RequestsTimeout = 500 * time.Millisecond
