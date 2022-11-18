package settings

import (
	"time"

	"github.com/go-chi/jwtauth"
	
)



// длинна укороченной ссылки без первого слеш
const KeyLeght int = 5 //значение должно быть больше 0

// длинна слайса байт для UserID
const UserIDLeght int = 16 //значение должно быть больше 0

// ключ подписи
const SignKey string = "9e9e0b4e6de418b2f84fca35165571c5"

// timeout запроса
const StorageTimeout = 60 * time.Second

// количество каналов для воркеров при установке пометку удаленный для sh_urls
const WorkersCount = 30

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
var RequestsTimeout = 0 * time.Second


