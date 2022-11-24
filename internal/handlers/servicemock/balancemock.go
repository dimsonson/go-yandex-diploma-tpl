package servicemock

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type Balance struct {
}

func (mserv *Balance) Status(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	return ec, nil
}
func (mserv *Balance) NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	return nil
}
func (mserv *Balance) WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	return ec, nil
}
