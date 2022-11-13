package services

import (
	"crypto/sha256"
	"fmt"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

// интерфейс методов хранилища
type StorageProvider interface {
	StorageConnectionClose()
}

// структура конструктора бизнес логики
type Services struct {
	storage StorageProvider
}

// конструктор бизнес логики
func NewService(s StorageProvider) *Services {
	return &Services{
		s,
	}
}

func (sr *Services) ServiceCreateNewUser(dc models.DecodeLoginPair) (err error) {

	src := []byte("Здесь могло быть написано, чем Go лучше Rust. " +
		"Но после хеширования уже не прочитаешь.")
	h := sha256.New()
	h.Write(src)
	dst := h.Sum(nil)
	fmt.Printf("%x", dst)

	return err
}

func (sr *Services) ServiceAuthorizationCheck(dc models.DecodeLoginPair) (err error) {
	return err
}

func (sr *Services) ServiceNewOrderLoad(login string, order_num string) (err error) {
	return err
}

// получение спска размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (sr *Services) ServiceGetOrdersList(login string) (ec models.OrdersList, err error) {
	return
}
