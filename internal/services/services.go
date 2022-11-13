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

	return err
}

func (sr *Services) ServiceAuthorizationCheck(dc models.DecodeLoginPair) (err error) {
	return err
}

// сервис загрузки пользователем номера заказа для расчёта
func (sr *Services) ServiceNewOrderLoad(login string, order_num string) (err error) {
	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (sr *Services) ServiceGetOrdersList(login string) (ec models.OrdersList, err error) {
	return
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (sr *Services) ServiceGetUserBalance(login string) (ec models.LoginBalance, err error) {
	return
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (sr *Services) ServiceNewWithdrawal(login string, dc models.NewWithdrawal) (err error) {
	return
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (sr *Services) ServiceGetWithdrawalsList(login string) (ec models.WithdrawalsList, err error) {
	return
}

func ToHex(src []byte) (dst []byte, err error) {
	//src = []byte("Здесь могло быть написано, чем Go лучше Rust. " +  "Но после хеширования уже не прочитаешь.")
	h := sha256.New()
	h.Write(src)
	dst = h.Sum(nil)
	fmt.Printf("%x", dst)
	return dst, err
}
