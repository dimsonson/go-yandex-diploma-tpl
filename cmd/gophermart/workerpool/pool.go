package workerpool

import (
	"context"
	"fmt"
	"sync"
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
	mu            sync.Mutex
	storage       StorageProvider
	calcSys       string
}

// канал остановки Pool
var done = make(chan struct{})

// конструктор задачи для воркера
func NewTask(LinkUpd string, Login string) *models.Task {
	return &models.Task{
		LinkUpd: LinkUpd,
		Login:   Login,
	}
}

// NewPool инициализирует новый пул с заданными задачами и при заданном параллелизме
func NewPool(tasks deque.Deque[models.Task], concurrency int, timeout *time.Ticker, storage StorageProvider, calcSys string) *Pool {
	return &Pool{
		TasksQ:      tasks,
		concurrency: concurrency,
		collector:   make(chan models.Task, settings.PipelineLenght),
		timeout:     timeout,
		storage:     storage,
		calcSys:     calcSys,
	}
}

// AddTask добавляет таски в pool
func (p *Pool) AppendTask(ctx context.Context, login, orderNum string) {
	// создаем ссылку для обноления статуса начислений по заказу
	linkUpd := fmt.Sprintf("%s/api/orders/%s", p.calcSys, orderNum)
	// создаем структуру для передачи в пул воркерам
	task := models.Task{
		Ctx:     ctx,
		LinkUpd: linkUpd,
		Login:   login,
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TasksQ.PushBack(task)
}

// RunBackground запускает pool в фоне
func (p *Pool) RunBackground() {
	// запуск воркеров с каналами получвения задач
	log.Print("starting Pool")
	for i := 1; i <= p.concurrency; i++ {
		worker := NewWorker(p.collector, i, p.timeout, p.storage)
		p.Workers = append(p.Workers, worker)
		go worker.StartBackground()
	}
	// передача задач из очереди в каналы воркеров
	for {
		select {
		// остановка пула
		case <-done:
			log.Print("closing Pool")
			return
		default:
			if lenQ := p.TasksQ.Len(); lenQ > 0 {
				p.mu.Lock()
				p.collector <- p.TasksQ.PopFront()
				p.mu.Unlock()
			}
		}
	}
}

// Stop останавливает запущенных в фоне воркеров
func (p *Pool) Stop() {
	defer close(done)
	for i := range p.Workers {
		p.Workers[i].Stop()
	}
}

// интерфейс доступа к хранилищу
type StorageProvider interface {
	Load(ctx context.Context, login string, orderNum string) (err error)
	List(ctx context.Context, login string) (ec []models.OrdersList, err error)
	Update(ctx context.Context, login string, dc models.OrderSatus) (err error)
}
