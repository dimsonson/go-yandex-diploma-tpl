package mocks

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type MockService struct {
}

func (m *MockService) ServiceCreateNewUser(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	return nil
}
func (m *MockService) ServiceAuthorizationCheck(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	return nil
}
func (m *MockService) ServiceNewOrderLoad(ctx context.Context, login string, orderNum string) (err error) {
	return nil
}
func (m *MockService) ServiceGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	ec = make([]models.OrdersList, 0)
	return ec, nil
}
func (m *MockService) ServiceGetUserBalance(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	ec = models.LoginBalance{}
	return ec, nil
}
func (m *MockService) ServiceNewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	return nil
}
func (m *MockService) ServiceGetWithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	ec = make([]models.WithdrawalsList, 0)
	return ec, nil
}
