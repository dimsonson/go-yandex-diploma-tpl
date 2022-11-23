package services

import (
	"context"
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

// интерфейс методов хранилища для Order
type Order interface {
	Load(ctx context.Context, login string, orderNum string) (err error)
	Update(ctx context.Context, login string, dc models.OrderSatus) (err error)
	List(ctx context.Context, login string) (ec []models.OrdersList, err error)
}

// структура конструктора бизнес логики Order
type OrderService struct {
	Order   Order
	CalcSys string
}

// конструктор бизнес логики Order
func NewOrderService(oStorage Order, calcSys string) *OrderService {
	return &OrderService{
		oStorage,
		calcSys,
	}
}

// сервис загрузки пользователем номера заказа для расчёта
func (storage *OrderService) Load(ctx context.Context, login string, orderNum string) (err error) {
	// проверка up and running внешнего сервиса
	rsp, err := http.Get(storage.CalcSys)
	if err != nil {
		log.Printf("remoute service error: %s", err)
		return
	}
	defer rsp.Body.Close()
	// запись нового заказа в хранилище
	err = storage.Order.Load(ctx, login, orderNum)
	if err != nil {
		return err
	}
	// создание ссылки для запроса в систему начисления баллов
	insertLink := storage.CalcSys + "/api/orders"
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
			linkUpd := fmt.Sprintf("%s/api/orders/%s", storage.CalcSys, orderNum)
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

				err = storage.Order.Update(ctx, login, dc)
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
func (storage *OrderService) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	ec, err = storage.Order.List(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}
