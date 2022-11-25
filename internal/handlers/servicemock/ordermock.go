package servicemock

import (
	"context"
	"errors"
	"log"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type Order struct {
}

func (mserv *Order) Load(ctx context.Context, login string, orderNum string) (err error) {
	switch {
	case login == "dimma" && orderNum == "1235489802":
		return nil
	case login == "dimma2login":
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
	switch {
case login == "dimma":
	return nil
case login == "dimma2login":
	log.Printf("order number from this login already exist: %s", orderNum)
	return errors.New("order number from this login already exist")
case login == "dimma3" :
	log.Printf("the same order number was loaded by another customer: %s", login)
	return errors.New("the same order number was loaded by another customer")
default:
	log.Printf("error for login: %s", login)
	return errors.New("something wrong woth server")
}
}
