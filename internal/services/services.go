package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	
)

// интерфейс методов хранилища
type StorageProvider interface {
	StorageConnectionClose()
	StorageCreateNewUser(ctx context.Context, login string, passwH string) (err error)
	StorageAuthorizationCheck(ctx context.Context, login string, passwHex string) (err error)
	StorageNewOrderLoad(ctx context.Context, login string, orderNum string) (err error)
	StorageGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error)
	StorageGetUserBalance(ctx context.Context, login string) (ec models.LoginBalance, err error)
	StorageNewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error)
	StorageGetWithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error)
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
func (sr *Services) ServiceNewOrderLoad(ctx context.Context, login string, orderNum string) (err error) {

	err = sr.storage.StorageNewOrderLoad(ctx, login, orderNum)
	if err != nil {
		log.Println(err)
		return err
	}

	b := fmt.Sprintf("{\"order\":\"%s\"}", orderNum)
	fmt.Println("b string", b)

	bPost, err := http.Post(sr.calcSys+"/api/orders", "application/json", strings.NewReader(b))
	if err != nil {
		log.Println("http.Post:", bPost, err)
		return err
	}
	defer bPost.Body.Close()

	fmt.Println("http.Post:", bPost, err) // ********

	bytes, err := io.ReadAll(bPost.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("http.Post Body:", string(bytes), err) // *****


	go func() {

		for {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			// пауза
			time.Sleep(settings.RequestsTimeout)

			link := fmt.Sprintf("%s/api/orders/%s", sr.calcSys, orderNum)

			fmt.Println(" ServiceNewOrderLoad link ::: ", link) //**********

			rGet, err := http.Get(link)
			if err != nil {
				log.Println("http.Get error :", err)
				return
			}

			fmt.Println("rGet:::", rGet)

			
			dc := models.OrderSatus{}
			// выполняем дальше, если нет 429 кода ответа
			if rGet.StatusCode != 429 {
				// десериализация тела ответа системы
				err = json.NewDecoder(rGet.Body).Decode(&dc) //`{"order":"521233510","status":"PROCESSED","accrual":729.98}`)).Decode(&dc) //rGet.Body).Decode(&dc) . //strings.NewReader("")
				if err != nil {
					log.Printf("unmarshal error ServiceNewOrderLoad gorutine: %s", err)
					return
				}

				err = sr.storage.StorageNewOrderUpdate(ctx, login, dc)
				if err != nil {
					log.Println("sr.storage.StorageNewOrderUpdate error :", err)
					return
				}
				// логируем
				log.Printf("login %s update order %s status to %s with accrual %v", login, dc.Order, dc.Status, dc.Accrual)
				if dc.Status == "INVALID" || dc.Status == "PROCESSED" {
					log.Printf("order %s has updated status to %s", dc.Order, dc.Status)
					return
				}
			}
			// пауза
			// time.Sleep(settings.RequestsTimeout)

			if rGet.StatusCode != 429 {
				timeout, err := strconv.Atoi(rGet.Header.Get("Retry-After"))
				if err != nil {
					log.Println("error converting Retry-After to int:", err)
					return
				}
				// увеличиваем паузу в соотвествии с Retry-After
				settings.RequestsTimeout = time.Duration(timeout) * time.Second
			}
			defer rGet.Body.Close()
		}

	}()

	log.Println("ServiceNewOrderLoad login, orderNum ", login, orderNum)

	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (sr *Services) ServiceGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	fmt.Println("ServiceGetOrdersList login", login, )
	ec, err = sr.storage.StorageGetOrdersList(ctx, login)
	fmt.Println("ServiceGetOrdersList login, ec :::", login, ec)
	return ec, err
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (sr *Services) ServiceGetUserBalance(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	fmt.Println("ServiceGetUserBalance login", login)
	ec, err = sr.storage.StorageGetUserBalance(ctx, login)
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (sr *Services) ServiceNewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	fmt.Println("ServiceNewWithdrawal login, dc", login, dc)
	err = sr.storage.StorageNewWithdrawal(ctx, login, dc)
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (sr *Services) ServiceGetWithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	fmt.Println("ServiceGetWithdrawalsList login", login)
	ec, err = sr.storage.StorageGetWithdrawalsList(ctx, login)
	return ec, err
}

// функция SHA.256 хеширования строки и кодирования хеша в строку
func ToHex(src string) (dst string, err error) {
	h := sha256.New()
	h.Write([]byte(src))
	tmpBytes := h.Sum(nil)
	dst = hex.EncodeToString(tmpBytes)
	return dst, err
}
