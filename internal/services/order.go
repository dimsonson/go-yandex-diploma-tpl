package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"

	"github.com/rs/zerolog/log"
)

// интерфейс методов хранилища для Order
type OrderStorageProvider interface {
	Load(ctx context.Context, login string, orderNum string) (err error)
	Update(ctx context.Context, login string, dc models.OrderSatus) (err error)
	List(ctx context.Context, login string) (ec []models.OrdersList, err error)
}

type PoolProvider interface {
	RunBackground()
	Stop()
	AppendTask(models.Task)
}

// структура конструктора бизнес логики Order
type OrderService struct {
	storage OrderStorageProvider
	CalcSys string
	pool    PoolProvider
}

// конструктор бизнес логики Order
func NewOrderService(oStorage OrderStorageProvider, calcSys string, pool PoolProvider) *OrderService {
	return &OrderService{
		oStorage,
		calcSys,
		pool,
	}
}

// сервис загрузки пользователем в систему начисления баллов номера нового заказа для расчёта
func (svc *OrderService) Load(ctx context.Context, login string, orderNum string) (err error) {
	// проверка up and running внешнего сервиса
	rsp, err := http.Get(svc.CalcSys)
	if err != nil {
		log.Printf("remoute service error: %s", err)
		return
	}
	defer rsp.Body.Close()
	// запись нового заказа в хранилище
	err = svc.storage.Load(ctx, login, orderNum)
	if err != nil {
		return err
	}
	// создание ссылки для запроса в систему начисления баллов
	insertLink := svc.CalcSys + "/api/orders"
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

	// создаем ссылку для обноления статуса начислений по заказу
	linkUpd := fmt.Sprintf("%s/api/orders/%s", svc.CalcSys, orderNum)

	task := &models.Task{
		LinkUpd: linkUpd,
		Login:   login,
	}
	svc.pool.AppendTask(*task)
	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (svc *OrderService) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	ec, err = svc.storage.List(ctx, login)
	// возвращаем структуру и ошибку
	return ec, err
}
