package storagemock

import (
	"context"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type Order struct {
}

func (mst *Order) Load(ctx context.Context, login string, orderNum string) (err error) {
	return nil
}
func (mst *Order) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	return ec, nil
}

func (mst *Order) Update(ctx context.Context, login string, dc models.OrderSatus) (err error){
	return nil
}