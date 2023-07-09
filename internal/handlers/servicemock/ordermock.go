package servicemock

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/shopspring/decimal"
)

// имплементация интерфейса OrderServiceProvider
type OrderServiceMock struct {
}

// заглушка
func (mserv *OrderServiceMock) Load(ctx context.Context, login string, orderNum string) (err error) {
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

// заглушка
func (mserv *OrderServiceMock) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	switch login {
	case "dimma":
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
	case "dimma2login":
		log.Printf("no orders for this login: %s", login)
		return nil, errors.New("no orders for this login")
	default:
		log.Printf("error for login: %s", login)
		return nil, errors.New("something wrong woth server")
	}
}
