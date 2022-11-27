package storagemock

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type Balance struct {
}

func (mst *Balance) Status(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	return ec, nil
}
func (mst *Balance) NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	return nil
}
func (mst *Balance) WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	return ec, nil
}
