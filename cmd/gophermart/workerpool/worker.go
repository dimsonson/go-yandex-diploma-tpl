package workerpool

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/rs/zerolog/log"
)

// структура воркера
type Worker struct {
	ID       int
	taskChan chan models.Task
	quit     chan bool
	timeoutW *time.Ticker
	storage  StorageProvider
}

// конструктор экземпляра воркера
func NewWorker(channel chan models.Task, ID int, timeout *time.Ticker, storage StorageProvider) *Worker {
	return &Worker{
		ID:       ID,
		taskChan: channel,
		quit:     make(chan bool),
		timeoutW: timeout,
		storage:  storage,
	}
}

// запуск воркера с выполнением задач по тикеру
func (wr *Worker) StartBackground() {
	log.Printf("starting Worker %d", wr.ID)
	for {
		select {
		case <-wr.timeoutW.C:
			task := <-wr.taskChan
			log.Printf("work of Worker %v : %v", wr.ID, task.LinkUpd)
			wr.Job(task)
		case <-wr.quit:
			return
		}
	}
}

// остановка воркеров
func (wr *Worker) Stop() {
	log.Printf("closing Worker %d", wr.ID)
	go func() {
		wr.quit <- true
	}()
}

// метод выполнения задачи для воркера
func (wr *Worker) Job(task models.Task) {
	for {
		// переопередяляем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
		// освобождаем ресурс
		defer cancel()

		// отпарвляем запрос на получения обновленных данных по заказу
		rGet, err := http.Get(task.LinkUpd)
		if err != nil {
			log.Printf("gorutine http Get error :%s", err)
			return
		}
		// завершаем задачу, если ордера нет в системе расчета баллов лояльности или заказ уже рассчитан
		if rGet.StatusCode == http.StatusNoContent || rGet.StatusCode == http.StatusConflict {
			log.Printf("status code %v recieved from extenal calculation service", rGet.StatusCode)
			return
		}
		// логгируем полученный статус заказа
		log.Printf("status code %v recieved from extenal calculation service", rGet.StatusCode)
		// выполняем дальше, если 200 код ответа
		if rGet.StatusCode == http.StatusOK {
			// десериализация тела ответа системы
			dc := models.OrderSatus{}
			err = json.NewDecoder(rGet.Body).Decode(&dc)
			if err != nil {
				log.Printf("unmarshal error ServiceNewOrderLoad gorutine: %s", err)
				return
			}
			// обновляем статус ордера в хранилище
			err = wr.storage.Update(ctx, task.Login, dc)
			if err != nil {
				log.Printf("sr.storage.StorageNewOrderUpdate error :%s", err)
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
		// если приходит 429 код ответа, увеличиваем таймаут запросов на значение в Retry-After
		if rGet.StatusCode == http.StatusTooManyRequests {
			timeout, err := strconv.Atoi(rGet.Header.Get("Retry-After"))
			if err != nil {
				log.Printf("error converting Retry-After to int:%s", err)
				return
			}
			// увеличиваем паузу в соотвествии с Retry-After
			wr.timeoutW = time.NewTicker(time.Duration(timeout) * 1000 * time.Millisecond)
		}
		// закрываем ресурс
		defer rGet.Body.Close()
	}
}