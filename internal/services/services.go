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
	fmt.Println("ServiceCreateNewUser dc", dc)
	return err
}

func (sr *Services) ServiceAuthorizationCheck(dc models.DecodeLoginPair) (err error) {
	fmt.Println("ServiceAuthorizationCheck dc", dc)
	return err
}

// сервис загрузки пользователем номера заказа для расчёта
func (sr *Services) ServiceNewOrderLoad(login string, order_num string) (err error) {
	fmt.Println("ServiceNewOrderLoad login, order_num ", login, order_num)
	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (sr *Services) ServiceGetOrdersList(login string) (ec models.OrdersList, err error) {
	fmt.Println("ServiceGetOrdersList login", login)
	ec = models.OrdersList{
		{
			Number:  "9278923470",
			Status:  "PROCESSED",
			Accrual: 500,
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
	return ec, err
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (sr *Services) ServiceGetUserBalance(login string) (ec models.LoginBalance, err error) {
	fmt.Println("ServiceGetUserBalance login", login)
	ec = models.LoginBalance{
		Current:   500.5,
		Withdrawn: 42,
	}
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (sr *Services) ServiceNewWithdrawal(login string, dc models.NewWithdrawal) (err error) {
	fmt.Println("ServiceNewWithdrawal login, dc", login, dc)
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (sr *Services) ServiceGetWithdrawalsList(login string) (ec models.WithdrawalsList, err error) {
	fmt.Println("ServiceGetWithdrawalsList login", login)
	ec = models.WithdrawalsList{
		{
			Order: "2377225624",
			Sum:   500,
			//ProcessedAt: "2020-12-09T16:09:57+03:00",
		},
		{
			Order: "2377225625",
			Sum:   800,
			//ProcessedAt: "2020-12-09T16:09:57+03:00",
		},
	}
	return ec, err
}

func ToHex(src []byte) (dst []byte, err error) {
	//src = []byte("Здесь могло быть написано, чем Go лучше Rust. " +  "Но после хеширования уже не прочитаешь.")
	h := sha256.New()
	h.Write(src)
	dst = h.Sum(nil)
	fmt.Printf("%x", dst)
	return dst, err
}
