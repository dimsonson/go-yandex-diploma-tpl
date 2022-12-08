package services

import (
	"context"
	"net/http"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"

	"github.com/rs/zerolog/log"
)

// интерфейс методов хранилища для Order
type OrderStorageProvider interface {
	Load(ctx context.Context, login string, orderNum string) (err error)
	List(ctx context.Context, login string) (ec []models.OrdersList, err error)
}

type PoolProvider interface {
	AppendTask(login, orderNum string)
}

type RequestProvider interface {
	RequestGet(url string) (rsp *http.Response, err error)
	RequestPost(orderNum string) (err error)
}

// структура конструктора бизнес логики Order
type OrderService struct {
	storage     OrderStorageProvider
	pool        PoolProvider
	httprequest RequestProvider
}

// конструктор бизнес логики Order
func NewOrderService(orderStorage OrderStorageProvider, pool PoolProvider, httprequest RequestProvider) *OrderService {
	return &OrderService{
		orderStorage,
		pool,
		httprequest,
	}
}

// сервис загрузки пользователем в систему начисления баллов номера нового заказа для расчёта
func (svc *OrderService) Load(ctx context.Context, login string, orderNum string) (err error) {
	// проверка up and running внешнего сервиса
	_, err = svc.httprequest.RequestGet("")
	if err != nil {
		log.Printf("remoute service request error (from OrderService Load): %s", err)
		return err
	}
	// запись нового заказа в хранилище
	err = svc.storage.Load(ctx, login, orderNum)
	if err != nil {
		return err
	}
	// запрос регистрации заказа в системе расчета баллов
	err = svc.httprequest.RequestPost(orderNum)
	if err != nil {
		log.Printf("http Post request in ServiceNewOrderLoad error:%s", err)
		return err
	}
	// отпарвляем запрос в пул воркеров для обработки
	svc.pool.AppendTask(login, orderNum)
	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (svc *OrderService) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	ec, err = svc.storage.List(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}
