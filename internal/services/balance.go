package services

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

// интерфейс методов хранилища для Balance
type Balance interface {
	NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error)
	WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error)
	Status(ctx context.Context, login string) (ec models.LoginBalance, err error)
}

// структура конструктора бизнес логики Balance
type BalanceService struct {
	storage Balance
}

// конструктор бизнес логики Balance
func NewBalanceService(bStorage Balance) *BalanceService {
	return &BalanceService{
		bStorage,
	}
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (storage *BalanceService) Status(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	ec, err = storage.storage.Status(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (storage *BalanceService) NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	err = storage.storage.NewWithdrawal(ctx, login, dc)
	// возвращаем ошибку
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (storage *BalanceService) WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	ec, err = storage.storage.WithdrawalsList(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}
