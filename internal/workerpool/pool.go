// пакет набора горутин-обработчиков для задач, с очередью обработки и закрытием по сигналу прерывания
package workerpool

import (
	"container/list"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"

	//"github.com/gammazero/deque"
	"github.com/rs/zerolog/log"
)

// структура пула воркеров
type Pool struct {
	TasksQ        *list.List // deque.Deque[models.Task]
	Workers       []*Worker
	concurrency   int
	collector     chan models.Task
	runBackground chan bool
	task          *models.Task
	timeout       *time.Ticker
	mu            sync.Mutex
	storage       StorageProvider
	calcSys       string
	wg            *sync.WaitGroup
	httprequest   HttpRequestProvider
}


// NewTask - конструктор структуры задач для воркера
func NewTask(orderNum string, Login string) *models.Task {
	return &models.Task{
		OrderNum: orderNum,
		Login:    Login,
	}
}

// NewPool инициализирует новый пул с заданными задачами и при заданном параллелизме
func NewPool(tasks *list.List /*deque.Deque[models.Task]*/, concurrency int, timeout *time.Ticker, storage StorageProvider, calcSys string, wg *sync.WaitGroup, httprequest HttpRequestProvider) *Pool {
	return &Pool{
		TasksQ:      tasks, //list.New(),
		concurrency: concurrency,
		collector:   make(chan models.Task, settings.PipelineLenght),
		timeout:     timeout,
		storage:     storage,
		calcSys:     calcSys,
		wg:          wg,
		httprequest: httprequest,
	}
}

// AppendTask добавляет задачи в pool
func (p *Pool) AppendTask(login, orderNum string) {
	// TaskQ:= list.New()
	// создаем структуру для передачи в очередь пула воркеров
	task := models.Task{
		OrderNum: orderNum,
		Login:    login,
	}
	// используем мьютексы для многопоточной работы с очередью
	p.mu.Lock()
	defer p.mu.Unlock()
	// добавлем задачу в конец очереди
	p.TasksQ.PushBack(task)
}

// RunBackground запускает пул воркеров
func (p *Pool) RunBackground(ctx context.Context) {
	log.Print("starting Pool")
	// запуск воркеров с каналами получения задач
	for i := 1; i <= p.concurrency; i++ {
		// констуруируем воркер
		worker := NewWorker(p.collector, i, p.timeout, p.storage, p.wg, p.httprequest)
		// добавляем воркер в слайс воркеров
		p.Workers = append(p.Workers, worker)
		// увеличиваем счетчик запущенных горутин
		p.wg.Add(1)
		// запускаем воркер
		go worker.StartBackground(ctx)
	}
	// передача задач из очереди в каналы воркеров
	for {
		select {
		// остановка пула по сигналу контекста
		case <-ctx.Done():
			log.Print("closing Pool")
			// уменьшем счетчик запущенных горутин
			p.wg.Done()
			return
		default:
			// если очередь не пустая, берем из нее задачу и отправляем в канал воркеров
			p.mu.Lock()
			if lenQ := p.TasksQ.Len(); lenQ > 0 {
				// удалем элемент из очереди с получением его значения и приведением его к типу models.Task с делаьнейшей передачей в канал
				p.collector <- p.TasksQ.Remove(p.TasksQ.Front()).(models.Task)
				//f := p.TasksQ.Front() //.Value.(models.Task)
				//element := p.TasksQ.Remove(p.TasksQ.Front()).(models.Task)
				//element
			}
			p.mu.Unlock()
		}
	}
}

// StorageProvider интерфейс доступа к хранилищу для методов пула воркеров
type StorageProvider interface {
	Update(ctx context.Context, login string, dc models.OrderSatus) (err error)
}

type HttpRequestProvider interface {
	RequestGet(url string) (rsp *http.Response, err error)
}
