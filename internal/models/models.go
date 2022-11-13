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

// баланс и сумма выводов средств по логину
type LoginBalance struct {
	Current   float64 `json:"current"`
	Withdrawn int     `json:"withdrawn"`
}

// структура списания с счета пользователя
type NewWithdrawal struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}

// информации о выводе средств с накопительного счёта пользователем
type WithdrawalsList []struct {
	Order       string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}