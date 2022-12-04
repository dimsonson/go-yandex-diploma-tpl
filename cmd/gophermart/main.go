package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/httprouter"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/storage"
	"github.com/gammazero/deque"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/dimsonson/go-yandex-diploma-tpl/cmd/gophermart/workerpool"
)

func init() {
	decimal.MarshalJSONWithoutQuotes = true
	decimal.DivisionPrecision = 2
	decimal.ExpMaxIterations = 1000
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006/01/02 15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	// получаем переменные
	dlink, calcSys, addr := flagsVars()
	// инициализируем конструкторы
	// конструкторы хранилища
	storage := newStrorageProvider(dlink)
	defer storage.ConnectionClose()
	// создаем тикер для воркер пула апдейта статусов заказов
	ticker := time.NewTicker(settings.RequestsTimeout)
	// создаем очередь для задач воркер пула апдейта статусов заказов
	queue := deque.New[models.Task]()
	// опередяляем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
	// освобождаем ресурс
	defer cancel()
	// создаем воркер пул апдейта статусов заказов
	pool := workerpool.NewPool(ctx, *queue, settings.WorkersQty, ticker, storage, calcSys)
	// конструкторы интерфейса User
	serviceUser := services.NewUserService(storage)
	handlerUser := handlers.NewUserHandler(serviceUser)
	//конструкторы интерфейса Order
	serviceOrder := services.NewOrderService(storage, calcSys, pool)
	handlerOrder := handlers.NewOrderHandler(serviceOrder)
	// конструкторы интерфейса Balance
	serviceBalance := services.NewBalanceService(storage)
	handlerBalance := handlers.NewBalanceHandler(serviceBalance)
	// конструктор роутера
	r := httprouter.NewRouter(handlerUser, handlerOrder, handlerBalance)
	// запускаем сервер
	log.Print("accruals calculation service URL: ", settings.ColorGreen, calcSys, settings.ColorReset)
	log.Print("starting http server on: ", settings.ColorBlue, addr, settings.ColorReset)
	// конфигурирование http сервера
	srv := &http.Server{Addr: addr, Handler: r}
	// канал остановки http сервера
	idleConnsClosed := make(chan struct{})
	// запуск http сервера ожидающего остановку
	go httpServerStart(srv, pool, idleConnsClosed)
	// запуск пула воркеров
	go pool.RunBackground()
	log.Print("ready to serve requests on " + addr)
	// получение сигнала остановки сервера и пула
	<-idleConnsClosed
	log.Print("gracefully shutting down")
}

// парсинг флагов и валидация переменных окружения
func flagsVars() (dlink string, calcSys string, addr string) {
	// описываем флаги
	addrFlag := flag.String("a", settings.DefServAddr, "HTTP Server address")
	calcSysFlag := flag.String("r", settings.DefCalcSysURL, "Accruals calculation service URL")
	dlinkFlag := flag.String("d", settings.DefDBlink, "Database URI link")
	// парсим флаги в переменные
	flag.Parse()
	// проверяем наличие переменной окружения, если ее нет или она не валидна, то используем значение из флага
	addr, ok := os.LookupEnv("RUN_ADDRESS")
	if !ok || !govalidator.IsURL(addr) || addr == "" {
		log.Print("eviroment variable RUN_ADDRESS is empty or has wrong value ", addr)
		addr = *addrFlag
	}
	// проверяем наличие переменной окружения, если ее нет или она не валидна, то используем значение из флага
	calcSys, ok = os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if !ok || !govalidator.IsURL(calcSys) || calcSys == "" {
		log.Print("eviroment variable ACCRUAL_SYSTEM_ADDRESS is empty or has wrong value ", calcSys)
		calcSys = *calcSysFlag
	}
	// проверяем наличие переменной окружения, если ее нет или она не валидна, то используем значение из флага
	dlink, ok = os.LookupEnv("DATABASE_URI")
	if !ok {
		log.Print("eviroment variable DATABASE_URI is not exist", dlink)
		dlink = *dlinkFlag
	}
	return dlink, calcSys, addr
}

// создание интерфейса хранилища
func newStrorageProvider(dlink string) (s *storage.StorageSQL) {
	// проверяем если переменная SQL url не пустая, логгируем
	if dlink == "" {
		log.Print("server may not properly start with "+settings.ColorRed+"database DSN empty", settings.ColorReset)
	}
	s = storage.NewSQLStorage(dlink)
	log.Print("server will start with data storage "+settings.ColorYellow+"in PostgreSQL:", dlink, settings.ColorReset)
	return s
}

// запуск и gracefull завершение ListenAndServe
func httpServerStart(srv *http.Server, pool *workerpool.Pool, idleConnsClosed chan struct{}) {
	go func() {
		// канал инциализирущего сигнала остановки сервера и пула воркеров
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		// сигнал получен, останавливаем воркеры
		pool.Stop()
		// сигнал получен, останавливаем сервер
		if err := srv.Shutdown(context.Background()); err != nil {
			// обработа ошибки остановки сервера
			log.Printf("HTTP server Shutdown error: %v", err)
		}
		// закрытие управляющего канала установки
		close(idleConnsClosed)
	}()
	// запуск http сервера
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// обработка ошибки запуска сервера
		log.Fatal().Err(err).Msgf("HTTP server ListenAndServe error: %v", err)
	}
}
