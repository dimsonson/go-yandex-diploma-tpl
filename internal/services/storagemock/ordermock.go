package storagemock

import (
	"context"
	"errors"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/shopspring/decimal"
)

type Order struct {
}

func (mst *Order) Load(ctx context.Context, login string, orderNum string) (err error) {

	if login == "dimma" && orderNum == "2377225624" {
		return nil
	}
	err = errors.New("something wrong with server")
	return err
}

func (mst *Order) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	if login == "dimma" {

		ec = []models.OrdersList{
			{
				Number:     "9278923470",
				Status:     "PROCESSED",
				Accrual:    decimal.NewFromFloatWithExponent(500, -2),
				UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
			},
			{
				Number:     "12345678903",
				Status:     "PROCESSING",
				UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
			},
			{
				Number:     "346436439",
				Status:     "INVALID",
				UploadedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
			},
		}
		return ec, nil
	}
	err = errors.New("something wrong with server")
	return nil, err
}
