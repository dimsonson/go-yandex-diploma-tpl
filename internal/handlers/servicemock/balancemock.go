package servicemock

import (
	"context"
	"errors"
	"log"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/shopspring/decimal"
)

type Balance struct {
}

func (mserv *Balance) Status(ctx context.Context, login string) (ec models.LoginBalance, err error) {
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

func (mserv *Balance) NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {

	switch {
	case login == "dimma" && dc.Order == "2377225624" && dc.Sum.Equal(decimal.NewFromFloat(751)):
		return nil
	case login == "dimma" && dc.Order == "2377225624" && dc.Sum.GreaterThan(decimal.NewFromFloat(751)):
		return errors.New("insufficient funds")
	case login == "dimma" && dc.Order == "24564564536456" :
		return errors.New("new order number already exist")
	default:
		log.Printf("error for login: %s", login)
		return errors.New("something wrong with server")
	}

}

func (mserv *Balance) WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	return ec, nil
}
