// API Накопительная система лояльности «Гофермарт»
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	// настройка обработки денежных единиц  
	decimal.MarshalJSONWithoutQuotes = true
	decimal.DivisionPrecision = 2
	decimal.ExpMaxIterations = 1000
	// настройка логгирования
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006/01/02 15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	// получаем переменные из флагов
	dlink, calcSys, addr := flagsVars()
	// инициализируем конструкторы
	// конструкторы хранилища
	storage := newStrorageProvider(dlink)
	defer storage.ConnectionClose()
	// создаем тикер для обработки задач из очереди
	ticker := time.NewTicker(settings.RequestsTimeout)
	// создаем очередь для задач воркер пула апдейта статусов заказов
	queue := deque.New[models.Task]()
	// опередяляем контекст уведомления о сигнале прерывания
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	// создаем группу синхранизации выполнения горутин
	var wg sync.WaitGroup
	// создаем воркер пул для обработки задач очереди
	pool := workerpool.NewPool(ctx, *queue, settings.WorkersQty, ticker, storage, calcSys, &wg)
	// конструкторы структур User
	serviceUser := services.NewUserService(storage)
	handlerUser := handlers.NewUserHandler(serviceUser)
	//конструкторы структур Order
	serviceOrder := services.NewOrderService(storage, calcSys, pool)
	handlerOrder := handlers.NewOrderHandler(serviceOrder)
	// конструкторы структур Balance
	serviceBalance := services.NewBalanceService(storage)
	handlerBalance := handlers.NewBalanceHandler(serviceBalance)
	// конструктор роутера
	r := httprouter.NewRouter(handlerUser, handlerOrder, handlerBalance)
	// запускаем сервер
	log.Print("accruals calculation service URL: ", settings.ColorGreen, calcSys, settings.ColorReset)
	log.Print("starting http server on: ", settings.ColorBlue, addr, settings.ColorReset)
	// конфигурирование http сервера
	srv := &http.Server{Addr: addr, Handler: r}
	// добавляем счетчик горутины
	wg.Add(1)
	// запуск горутины shutdown http сервера
	go httpServerShutdown(ctx, &wg, srv)
	// добавляем счетчик горутины
	wg.Add(1)
	// запуск горутины пула воркеров
	go pool.RunBackground()
	// запуск http сервера
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// обработка ошибки запуска сервера
		log.Fatal().Err(err).Msgf("HTTP server ListenAndServe error: %v", err)
	}
	// ожидаем выполнение горутин
	wg.Wait()
	// остановка всех сущностей, куда передан контекст по прерыванию
	stop()
	// логирование закрытия сервера без ошибок
	log.Print("http server gracefully shutdown")
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

// создание структуры хранилища
func newStrorageProvider(dlink string) (s *storage.StorageSQL) {
	// проверяем если переменная SQL url не пустая, логгируем
	if dlink == "" {
		log.Print("server may not properly start with "+settings.ColorRed+"database DSN empty", settings.ColorReset)
	}
	// создаем хранилище через конструктор
	s = storage.NewSQLStorage(dlink)
	log.Print("server will start with data storage "+settings.ColorYellow+"in PostgreSQL:", dlink, settings.ColorReset)
	return s
}

// gracefull shutdown для ListenAndServe
func httpServerShutdown(ctx context.Context, wg *sync.WaitGroup, srv *http.Server) {
	// получаем сигнал о завершении приложения
	<-ctx.Done()
	// завершаем открытые соединения и закрываем http server
	if err := srv.Shutdown(ctx); err != nil {
		// логирование ошибки остановки сервера
		log.Printf("HTTP server Shutdown error: %v", err)
	}
	// уменьшаем счетчик запущенных горутин
	wg.Done()
}
