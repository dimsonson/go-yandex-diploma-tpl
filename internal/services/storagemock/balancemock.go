package storagemock

import (
	"context"
	"errors"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/shopspring/decimal"
)

type Balance struct {
}

func (mst *Balance) Status(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	ec = models.LoginBalance{
		Current:   decimal.NewFromFloatWithExponent(500.505, -2),
		Withdrawn: decimal.NewFromFloatWithExponent(42, -2),
	}
	err = errors.New("something wrong woth server")
	return ec, err
}

func (mst *Balance) NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	dcCorr := models.NewWithdrawal{
		Order: "2377225624",
		Sum:   decimal.NewFromFloatWithExponent(42, -2),
	}
	if dc.Sum.Equal(dcCorr.Sum) && dc.Order == dcCorr.Order {
		return nil
	}
	err = errors.New("something wrong woth server")
	return err
}

func (mst *Balance) WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	if login == "dimma" {

		ec = []models.WithdrawalsList{
			{
				Order:       "2377225624",
				Sum:         decimal.NewFromFloatWithExponent(500.0300, -2),
				ProcessedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
			},
			{
				Order:       "2377225625",
				Sum:         decimal.NewFromFloatWithExponent(800.5555, -2),
				ProcessedAt: time.Date(2020, time.May, 15, 17, 45, 12, 0, time.Local),
			}}
		return ec, nil
	}
	err = errors.New("something wrong woth server")
	return nil, err
}
