package workerpool

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/rs/zerolog/log"
)

// Worker - структура воркера
type Worker struct {
	ID          int
	taskChan    chan models.Task
	quit        chan bool
	timeoutW    *time.Ticker
	storage     StorageProvider
	wg          *sync.WaitGroup
	httprequest HttpRequestProvider
}

// NewWorker - конструктор экземпляра воркера
func NewWorker(taskChan chan models.Task, ID int, timeout *time.Ticker, storage StorageProvider, wg *sync.WaitGroup, httprequest HttpRequestProvider) *Worker {
	return &Worker{
		ID:          ID,
		taskChan:    taskChan,
		quit:        make(chan bool),
		timeoutW:    timeout,
		storage:     storage,
		wg:          wg,
		httprequest: httprequest,
	}
}

// StartBackground запускает воркер с выполнением задач по тикеру для поддержаия RPM запросов
func (wr *Worker) StartBackground(ctx context.Context) {
	log.Printf("starting Worker %d", wr.ID)
	for {
		select {
		// получаем сигнал тикера
		case <-wr.timeoutW.C:
			// если канал получения задач не пустой, получем из него задачу и обрабатываем
			select {
			case task := <-wr.taskChan:
				log.Printf("work of Worker %v : %v", wr.ID, task.OrderNum)
				// запуск метода выполнения задачи
				wr.Job(ctx, task)
				// если канал с задачами пустой - ничего не делаем
			default:
			}
			// получаем сигнал оостановки
		case <-ctx.Done():
			log.Printf("closing Worker %d", wr.ID)
			// уменьшаем счетчик запущенных горутин
			wr.wg.Done()
			return
		}
	}
}

// Job - метод выполнения задачи для воркера
func (wr *Worker) Job(ctx context.Context, task models.Task) {
	for {
		// отпарвляем запрос в внешний сервис на получения обновленных данных по заказу
		rGet, err := wr.httprequest.RequestGet(task.OrderNum)
		if err != nil {
			log.Printf("gorutine http Get error :%s", err)
			return
		}
		// завершаем задачу, если ордера нет в системе расчета баллов лояльности или заказ уже рассчитан
		if rGet.StatusCode == http.StatusNoContent || rGet.StatusCode == http.StatusNotFound || rGet.StatusCode == http.StatusConflict {
			log.Printf("status code %v recieved from extenal calculation service", rGet.StatusCode)
			return
		}
		// логгируем полученный статус код ответа внешнего сервиса
		log.Printf("http status code %v recieved from extenal calculation service", rGet.StatusCode)
		// выполняем дальше, если 200 код ответа
		if rGet.StatusCode == http.StatusOK {
			// десериализация тела ответа системы
			dc := models.OrderSatus{}
			err = json.NewDecoder(rGet.Body).Decode(&dc)
			if err != nil {
				log.Printf("unmarshal error Worker Job gorutine: %s", err)
				return
			}
			// обновляем статус ордера в хранилище
			err = wr.storage.Update(ctx, task.Login, dc)
			if err != nil {
				log.Printf("storage.Update Worker Job error :%s", err)
				return
			}
			// логируем обновление в хранилище
			log.Printf("login %s update order %s status to %s with accrual %v", task.Login, dc.Order, dc.Status, dc.Accrual)
			// останавливаем задачу, если получен финальный стаус
			if dc.Status == "INVALID" || dc.Status == "PROCESSED" {
				log.Printf("order %s has updated status to %s", dc.Order, dc.Status)
				return
			}
		}
		// если приходит 429 код ответа, делаем паузу на значение в Retry-After
		if rGet.StatusCode == http.StatusTooManyRequests {
			timeout, err := strconv.Atoi(rGet.Header.Get("Retry-After"))
			if err != nil {
				log.Printf("error converting Retry-After to int:%s", err)
				return
			}
			// делаем паузу в соотвествии с Retry-After
			<-time.After(time.Duration(timeout) * time.Second)
		}
		// закрываем ресурс
		defer rGet.Body.Close()
	}
}
