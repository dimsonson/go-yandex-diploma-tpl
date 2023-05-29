package servicemock

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/shopspring/decimal"
)

// имплементация интерфейса BalanceServiceProvider
type BalanceServiceProvider struct {
}

// заглушка Status
func (mserv *BalanceServiceProvider) Status(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	switch login {

	case "dimma":
		ec = models.LoginBalance{
			Current:   decimal.NewFromFloatWithExponent(500.505, -2),
			Withdrawn: decimal.NewFromFloatWithExponent(42, -2),
		}
		return ec, nil

	default:
		log.Printf("error for login: %s", login)
		return ec, errors.New("something wrong with server")
	}

}

// заглушка NewWithdrawal
func (mserv *BalanceServiceProvider) NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {

	switch {
	case login == "dimma" && dc.Order == "2377225624" && dc.Sum.Equal(decimal.NewFromFloat(751)):
		return nil
	case login == "dimma" && dc.Order == "2377225624" && dc.Sum.GreaterThan(decimal.NewFromFloat(751)):
		return errors.New("insufficient funds")
	case login == "dimma" && dc.Order == "24564564536456":
		return errors.New("new order number already exist")
	default:
		log.Printf("error for login: %s", login)
		return errors.New("something wrong with server")
	}

}

// заглушка WithdrawalsList
func (mserv *BalanceServiceProvider) WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
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
		},
	}
	switch login {
	case "dimma":
		return ec, nil
	case "dimma2":
		return nil, errors.New("no records")
	default:
		log.Printf("error for login: %s", login)
		return nil, errors.New("something wrong with server")
	}
}
