package servicemock

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type Order struct {
}

func (mserv *Order) Load(ctx context.Context, login string, orderNum string) (err error) {
	return nil
}
func (mserv *Order) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	return ec, nil
}
