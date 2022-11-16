package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	//"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
)

// интерфейс методов хранилища
type StorageProvider interface {
	StorageConnectionClose()
	StorageCreateNewUser(ctx context.Context, login string, passwH string) (err error)
	StorageAuthorizationCheck(ctx context.Context, login string, passwHex string) (err error)
	StorageNewOrderLoad(ctx context.Context, login string, order_num string) (err error)
	StorageGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error)
	StorageGetUserBalance(login string) (ec models.LoginBalance, err error)
	StorageNewWithdrawal(login string, dc models.NewWithdrawal) (err error)
	StorageGetWithdrawalsList(login string) (ec models.WithdrawalsList, err error)
	StorageNewOrderUpdate(ctx context.Context, login string, dc models.OrderSatus) (err error)
}

// структура конструктора бизнес логики
type Services struct {
	storage StorageProvider
	calcSys string
}

// конструктор бизнес логики
func NewService(s StorageProvider, calcSys string) *Services {
	return &Services{
		s,
		calcSys,
	}
}

func (sr *Services) ServiceCreateNewUser(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	fmt.Println("ServiceCreateNewUser dc", dc)
	// сощдание хеш пароля для передачи в хранилище
	passwHex, err := ToHex(dc.Password)
	if err != nil {
		log.Println("hex conversion in ServiceCreateNewUser error :", err)
		return err
	}
	// передача пары логин:пароль в хранилище
	err = sr.storage.StorageCreateNewUser(ctx, dc.Login, passwHex)
	return err
}

func (sr *Services) ServiceAuthorizationCheck(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	fmt.Println("ServiceAuthorizationCheck dc", dc)
	// сощдание хеш пароля для передачи в хранилище
	passwHex, err := ToHex(dc.Password)
	if err != nil {
		log.Println("hex conversion in ServiceCreateNewUser error :", err)
		return err
	}
	// передача пары логин:пароль в хранилище
	err = sr.storage.StorageAuthorizationCheck(ctx, dc.Login, passwHex)
	return err
}

// сервис загрузки пользователем номера заказа для расчёта
func (sr *Services) ServiceNewOrderLoad(ctx context.Context, login string, order_num string) (err error) {
	b := fmt.Sprintf("{\"order\":\"%s\"}", order_num)
	fmt.Println("b string", b)

	bPost, err := http.Post(sr.calcSys, "application/json", strings.NewReader(b))
	if err != nil {
		log.Println("http.Post:", bPost, err)
		return err
	}
	fmt.Println("http.Post:", bPost, err)
	fmt.Println("http.Post Body:", bPost.Body, err)

	err = sr.storage.StorageNewOrderLoad(ctx, login, order_num)
	if err != nil {
		log.Println(err)
		return err
	}

	go func() {

		for {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			g := fmt.Sprintf("%s/%s", sr.calcSys, order_num)
			rGet, err := http.Get(g)
			if err != nil {
				log.Println("http.Get error :", err)
				return
			}

			// десериализация тела ответа системы
			dc := models.OrderSatus{}
			err = json.NewDecoder(rGet.Body).Decode(&dc)
			if err != nil {
				log.Printf("Unmarshal error: %s", err)
				return
			}

			err = sr.storage.StorageNewOrderUpdate(ctx, login, dc)
			if err != nil {
				log.Println("sr.storage.StorageNewOrderUpdate error :", err)
				return
			}

			log.Printf("login %s update order %s status to %s with accrual %v", login, dc.Order, dc.Status, dc.Accrual)

			if dc.Status == "INVALID" || dc.Status == "PROCESSED" {
				log.Printf("order %s has updated status to %s", dc.Order, dc.Status)
				return
			}

			fmt.Printf("dc:\n Accrual: %v\n Order: %v\n Status: %v\n", dc.Accrual, dc.Order, dc.Status)
			fmt.Println("http.Get:", rGet, err)
		}

	}()

	log.Println("ServiceNewOrderLoad login, order_num ", login, order_num)

	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (sr *Services) ServiceGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	fmt.Println("ServiceGetOrdersList login", login)
	ec, err = sr.storage.StorageGetOrdersList(ctx, login)
	return ec, err
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (sr *Services) ServiceGetUserBalance(login string) (ec models.LoginBalance, err error) {
	fmt.Println("ServiceGetUserBalance login", login)
	ec, err = sr.storage.StorageGetUserBalance(login)
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (sr *Services) ServiceNewWithdrawal(login string, dc models.NewWithdrawal) (err error) {
	fmt.Println("ServiceNewWithdrawal login, dc", login, dc)
	err = sr.storage.StorageNewWithdrawal(login, dc)
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (sr *Services) ServiceGetWithdrawalsList(login string) (ec models.WithdrawalsList, err error) {
	fmt.Println("ServiceGetWithdrawalsList login", login)
	ec, err = sr.storage.StorageGetWithdrawalsList(login)
	return ec, err
}

func ToHex(src string) (dst string, err error) {
	//src = []byte("Здесь могло быть написано, чем Go лучше Rust. " +  "Но после хеширования уже не прочитаешь.")
	h := sha256.New()
	h.Write([]byte(src))
	tmpBytes := h.Sum(nil)
	dst = hex.EncodeToString(tmpBytes)

	// fmt.Printf("%x\n", dst)
	return dst, err
}
