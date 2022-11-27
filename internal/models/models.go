package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// структура декодирования пары логин:пароль
type DecodeLoginPair struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// список заказов
type OrdersList struct {
	Number     string          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at"`
}

// баланс и сумма выводов средств по логину
type LoginBalance struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

// структура списания с счета пользователя
type NewWithdrawal struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}

// информации о выводе средств с накопительного счёта пользователем
type WithdrawalsList struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}

// статус ордера из истемы начислений
type OrderSatus struct {
	Order   string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

// задача для воркера работающего с внешним версиом начислений баллов лояльности
type Task struct {
	LinkUpd string
	Login string
	
}
