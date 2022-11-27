package servicemock

import (
	"context"
	"errors"
	"log"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/shopspring/decimal"
)

type Order struct {
}

func (mserv *Order) Load(ctx context.Context, login string, orderNum string) (err error) {
	switch {
	case login == "dimma" && orderNum == "1235489802":
		return nil
	case login == "dimma2login" && orderNum == "1235489802":
		log.Printf("order number from this login already exist: %s", orderNum)
		return errors.New("order number from this login already exist")
	case login == "dimma3" && orderNum == "1235489802":
		log.Printf("the same order number was loaded by another customer: %s", login)
		return errors.New("the same order number was loaded by another customer")
	default:
		log.Printf("error for login: %s", login)
		return errors.New("something wrong woth server")
	}

}
func (mserv *Order) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	switch login {
	case "dimma":
		ec = []models.OrdersList{
			{
				Number:  "9278923470",
				Status:  "PROCESSED",
				Accrual: decimal.NewFromFloatWithExponent(500, -2),
				// UploadedAt: "2020-12-10T15:15:45+03:00",
			},
			{
				Number: "12345678903",
				Status: "PROCESSING",
				// UploadedAt: "2020-12-10T15:12:01+03:00",
			},
			{
				Number: "346436439",
				Status: "INVALID",
				// UploadedAt: "2020-12-09T16:09:53+03:00",
			},
		}
		return ec, nil
	case "dimma2login":
		log.Printf("no orders for this login: %s", login)
		return nil, errors.New("no orders for this login")
	default:
		log.Printf("error for login: %s", login)
		return nil, errors.New("something wrong woth server")
	}
}
