// пакет набора горутин-обработчиков для задач, с очередью обработки и закрытием по сигналу прерывания
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
	ctx           context.Context
	wg            *sync.WaitGroup
}

// NewTask - конструктор структуры задач для воркера
func NewTask(LinkUpd string, Login string) *models.Task {
	return &models.Task{
		LinkUpd: LinkUpd,
		Login:   Login,
	}
}

// NewPool инициализирует новый пул с заданными задачами и при заданном параллелизме
func NewPool(ctx context.Context, tasks deque.Deque[models.Task], concurrency int, timeout *time.Ticker, storage StorageProvider, calcSys string, wg *sync.WaitGroup) *Pool {
	return &Pool{
		ctx:         ctx,
		TasksQ:      tasks,
		concurrency: concurrency,
		collector:   make(chan models.Task, settings.PipelineLenght),
		timeout:     timeout,
		storage:     storage,
		calcSys:     calcSys,
		wg:          wg,
	}
}

// AppendTask добавляет задачи в pool
func (p *Pool) AppendTask(login, orderNum string) {
	// создаем ссылку для обноления статуса начислений по заказу
	linkUpd := fmt.Sprintf("%s/api/orders/%s", p.calcSys, orderNum)
	// создаем структуру для передачи в очередь пула воркеров
	task := models.Task{
		LinkUpd: linkUpd,
		Login:   login,
	}
	// используем мьютексы для многопоточной работы с очередью
	p.mu.Lock()
	defer p.mu.Unlock()
	// добавлем задачу в конец очереди
	p.TasksQ.PushBack(task)
}

// RunBackground запускает пул воркеров
func (p *Pool) RunBackground() {
	log.Print("starting Pool")
	// запуск воркеров с каналами получения задач
	for i := 1; i <= p.concurrency; i++ {
		// констуруируем воркер
		worker := NewWorker(p.ctx, p.collector, i, p.timeout, p.storage, p.wg)
		// добавляем воркер в слайс воркеров
		p.Workers = append(p.Workers, worker)
		// увеличиваем счетчик запущенных горутин
		p.wg.Add(1)
		// запускаем воркер
		go worker.StartBackground()
	}
	// передача задач из очереди в каналы воркеров
	for {
		select {
		// остановка пула по сигналу контекста
		case <-p.ctx.Done():
			log.Print("closing Pool")
			// уменьшем счетчик запущенных горутин
			p.wg.Done()
			return
		default:
			// если очередь не пустая, берем из нее задачу и отправляем в канал воркеров
			p.mu.Lock()
			if lenQ := p.TasksQ.Len(); lenQ > 0 {
				p.collector <- p.TasksQ.PopFront()
			}
			p.mu.Unlock()
		}
	}
}

// StorageProvider интерфейс доступа к хранилищу для методов пула воркеров
type StorageProvider interface {
	Update(ctx context.Context, login string, dc models.OrderSatus) (err error)
}
