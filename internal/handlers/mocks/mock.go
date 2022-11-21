package mocks

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type Services struct {
}

func (m *Services) ServiceCreateNewUser(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	return nil
}
func (m *Services) ServiceAuthorizationCheck(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	return nil
}
func (m *Services) ServiceNewOrderLoad(ctx context.Context, login string, orderNum string) (err error) {
	return nil
}
func (m *Services) ServiceGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	ec = make([]models.OrdersList, 0)
	return ec, nil
}
func (m *Services) ServiceGetUserBalance(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	ec = models.LoginBalance{}
	return ec, nil
}
func (m *Services) ServiceNewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	return nil
}
func (m *Services) ServiceGetWithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	ec = make([]models.WithdrawalsList, 0)
	return ec, nil
}
