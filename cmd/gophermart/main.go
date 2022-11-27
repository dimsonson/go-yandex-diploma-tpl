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


// переменные по умолчанию
const (
	defServAddr   = "localhost:8000"
	defDBlink     = "postgres://postgres:1818@localhost:5432/gophm"
	defCalcSysURL = "http://localhost:8080"
)

func main() {
	// получаем переменные
	dlink, calcSys, addr := flagsVars()
	// инициализируем конструкторы
	// конструкторы хранилища
	storage := newStrorageProvider(dlink)
	defer storage.ConnectionClose()

	ticker := time.NewTicker(settings.RequestsTimeout)
	queue := deque.New[models.Task]()
	pool := workerpool.NewPool(*queue, settings.WorkersQty, ticker, storage)

	// конструкторы User
	serviceUser := services.NewUserService(storage)
	handlerUser := handlers.NewUserHandler(serviceUser)
	//конструкторы Order
	serviceOrder := services.NewOrderService(storage, calcSys, pool)
	handlerOrder := handlers.NewOrderHandler(serviceOrder)
	// конструкторы Balance
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
	// запуск сервера ожидающего остановку
	go httpServerStart(srv, pool, idleConnsClosed)

	go pool.RunBackground()

	log.Print("ready to serve requests on " + addr)
	// получение сигнала остановки
	<-idleConnsClosed
	log.Print("gracefully shutting down")
}

// парсинг флагов и валидация переменных окружения
func flagsVars() (dlink string, calcSys string, addr string) {
	// описываем флаги
	addrFlag := flag.String("a", defServAddr, "HTTP Server address")
	calcSysFlag := flag.String("r", defCalcSysURL, "Accruals calculation service URL")
	dlinkFlag := flag.String("d", defDBlink, "Database URI link")
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
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		// останавливаем воркеры
		pool.Stop()
		// we received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatal().Err(err).Msgf("HTTP server ListenAndServe error: %v", err)
	}
}
