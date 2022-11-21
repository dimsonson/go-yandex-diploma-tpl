package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"

	"github.com/rs/zerolog/log"
)

//go:generate mockgen -source=services.go -destination=mocks/mock.go

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
	// сощдание хеш пароля для передачи в хранилище
	passwHex, err := ToHex(dc.Password)
	if err != nil {
		log.Printf("hex conversion in ServiceCreateNewUser error :%s", err)
		return err
	}
	// передача пары логин:пароль в хранилище
	err = sr.storage.StorageCreateNewUser(ctx, dc.Login, passwHex)
	return err
}

func (sr *Services) ServiceAuthorizationCheck(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	// сощдание хеш пароля для передачи в хранилище
	passwHex, err := ToHex(dc.Password)
	if err != nil {
		log.Printf("hex conversion in ServiceCreateNewUser error :%s", err)
		return err
	}
	// передача пары логин:пароль в хранилище
	err = sr.storage.StorageAuthorizationCheck(ctx, dc.Login, passwHex)
	return err
}

// сервис загрузки пользователем номера заказа для расчёта
func (sr *Services) ServiceNewOrderLoad(ctx context.Context, login string, orderNum string) (err error) {
	// проверка up and running внешнего сервиса
	rsp, err := http.Get(sr.calcSys)
	if err != nil {
		log.Printf("remoute service error: %s", err)
		return
	}
	defer rsp.Body.Close()
	// запись нового заказа в хранилище
	err = sr.storage.StorageNewOrderLoad(ctx, login, orderNum)
	if err != nil {
		return err
	}
	// создание ссылки для запроса в систему начисления баллов
	insertLink := sr.calcSys + "/api/orders"
	// создание JSON для запроса в систему начисления баллов
	bodyJSON := fmt.Sprintf("{\"order\":\"%s\"}", orderNum)
	// запрос регистрации заказа в системе расчета баллов
	rPost, err := http.Post(insertLink, "application/json", strings.NewReader(bodyJSON))
	if err != nil {
		log.Printf("http Post request in ServiceNewOrderLoad error:%s", err)
		return err
	}
	// освобождаем ресурс
	defer rPost.Body.Close()
	// запуск горутины обновления статуса начсления баллов по заказу
	go func() {

		for {
			// переопередяляем контекст с таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			// пауза
			time.Sleep(settings.RequestsTimeout)
			// создаем ссылку для обноления статуса начислений по заказу
			linkUpd := fmt.Sprintf("%s/api/orders/%s", sr.calcSys, orderNum)
			// отпарвляем запрос на получения обновленных данных по заказу
			rGet, err := http.Get(linkUpd)
			if err != nil {
				log.Printf("gorutine http Get error :%s", err)
				return
			}
			// выполняем дальше, если 200 код ответа
			if rGet.StatusCode == 200 {
				// десериализация тела ответа системы
				dc := models.OrderSatus{}
				err = json.NewDecoder(rGet.Body).Decode(&dc)
				if err != nil {
					log.Printf("unmarshal error ServiceNewOrderLoad gorutine: %s", err)
					return
				}

				err = sr.storage.StorageNewOrderUpdate(ctx, login, dc)
				if err != nil {
					log.Printf("sr.storage.StorageNewOrderUpdate error :%s", err)
					return
				}
				// логируем
				log.Printf("login %s update order %s status to %s with accrual %v", login, dc.Order, dc.Status, dc.Accrual)
				if dc.Status == "INVALID" || dc.Status == "PROCESSED" {
					log.Printf("order %s has updated status to %s", dc.Order, dc.Status)
					return
				}
			}
			// если приходит 429 код ответа, увеличиваем таймаут запросов на значение в Retry-After
			if rGet.StatusCode == 429 {
				timeout, err := strconv.Atoi(rGet.Header.Get("Retry-After"))
				if err != nil {
					log.Printf("error converting Retry-After to int:%s", err)
					return
				}
				// увеличиваем паузу в соотвествии с Retry-After
				settings.RequestsTimeout = time.Duration(timeout) * 1000 * time.Millisecond
			}
			// закрываем ресурс
			defer rGet.Body.Close()
		}
	}()
	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (sr *Services) ServiceGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	ec, err = sr.storage.StorageGetOrdersList(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (sr *Services) ServiceGetUserBalance(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	ec, err = sr.storage.StorageGetUserBalance(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (sr *Services) ServiceNewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	err = sr.storage.StorageNewWithdrawal(ctx, login, dc)
	// возвращаем ошибку
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (sr *Services) ServiceGetWithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	ec, err = sr.storage.StorageGetWithdrawalsList(ctx, login)
	// возвращаем структуру и ошибку
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
