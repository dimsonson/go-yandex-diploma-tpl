package services

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

// интерфейс методов хранилища для Balance
type BalanceStorageProvider interface {
	NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error)
	WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error)
	Status(ctx context.Context, login string) (ec models.LoginBalance, err error)
}

// структура конструктора бизнес логики Balance
type BalanceService struct {
	storage BalanceStorageProvider
}

// конструктор бизнес логики Balance
func NewBalanceService(bStorage BalanceStorageProvider) *BalanceService {
	return &BalanceService{
		bStorage,
	}
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (svc *BalanceService) Status(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	ec, err = svc.storage.Status(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (svc *BalanceService) NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	err = svc.storage.NewWithdrawal(ctx, login, dc)
	// возвращаем ошибку
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (svc *BalanceService) WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	ec, err = svc.storage.WithdrawalsList(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}
