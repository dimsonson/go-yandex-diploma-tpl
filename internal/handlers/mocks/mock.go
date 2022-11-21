package mocks

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type Service struct {
}

func (m *Service) ServiceCreateNewUser(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	return nil
}
func (m *Service) ServiceAuthorizationCheck(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	return nil
}
func (m *Service) ServiceNewOrderLoad(ctx context.Context, login string, orderNum string) (err error) {
	return nil
}
func (m *Service) ServiceGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	ec = make([]models.OrdersList, 0)
	return ec, nil
}
func (m *Service) ServiceGetUserBalance(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	ec = models.LoginBalance{}
	return ec, nil
}
func (m *Service) ServiceNewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	return nil
}
func (m *Service) ServiceGetWithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	ec = make([]models.WithdrawalsList, 0)
	return ec, nil
}
