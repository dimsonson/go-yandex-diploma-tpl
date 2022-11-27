package workerpool

import (
	"context"
	"time"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/gammazero/deque"
	"github.com/rs/zerolog/log"
)

// структура пула воркеров
type Pool struct {
	TasksQ        deque.Deque[models.Task]
	Workers       []*Worker
	concurrency   int
	collector     chan models.Task
	runBackground chan bool
	task          *models.Task
	timeout       *time.Ticker
	storage       StorageProvider
}
// канал остановки Pool
var done = make(chan struct{})

// конструктор задачи для воркера
func NewTask(LinkUpd string, Login string) *models.Task {
	return &models.Task{
		LinkUpd: LinkUpd,
		Login:   Login,
		//service: service,
	}
}

// NewPool инициализирует новый пул с заданными задачами и при заданном параллелизме
func NewPool(tasks deque.Deque[models.Task], concurrency int, timeout *time.Ticker, storage StorageProvider) *Pool {
	return &Pool{
		TasksQ:      tasks,
		concurrency: concurrency,
		collector:   make(chan models.Task, settings.PipelineLenght),
		timeout:     timeout,
		storage:     storage,
	}
}

// AddTask добавляет таски в pool
func (p *Pool) AppendTask(task models.Task) {
	p.TasksQ.PushBack(task)
}

// RunBackground запускает pool в фоне
func (p *Pool) RunBackground() {
	log.Print("starting Pool")
	for i := 1; i <= p.concurrency; i++ {
		worker := NewWorker(p.collector, i, p.timeout, p.storage)
		p.Workers = append(p.Workers, worker)
		go worker.StartBackground()
	}
	p.runBackground = make(chan bool)
	for {
		select {
		case <-done:
			log.Print("closing Pool")
			return
		default:
			if lenQ := p.TasksQ.Len(); lenQ > 0 {
				p.collector <- p.TasksQ.PopFront()
			}
		}
	}
}

// Stop останавливает запущенных в фоне worker-ов
func (p *Pool) Stop() {
	defer close(done)
	for i := range p.Workers {
		p.Workers[i].Stop()
	}
}

// интерфейс доступа к хранилищу
type StorageProvider interface {
	Load(ctx context.Context, login string, orderNum string) (err error)
	Update(ctx context.Context, login string, dc models.OrderSatus) (err error)
	List(ctx context.Context, login string) (ec []models.OrdersList, err error)
}
