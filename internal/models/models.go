package models

import "time"

// структура декодирования пары логин:пароль
type DecodeLoginPair struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// список заказов
type OrdersList []struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
