package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/handlers"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/httprouter"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/services"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/storage"
)

// переменные по умолчанию
const (
	defServAddr   = "localhost:8080"
	defDBlink     = "postgres://postgres:1818@localhost:5432/gophm" // новую базу
	defCalcSysURL = "localhost:8080"
)

func main() {
	// получаем переменные
	dlink, calcSys, addr := flagsVars()
	// инициализируем конструкторы
	s := newStrorageProvider(dlink)
	defer s.StorageConnectionClose()
	srvs := services.NewService(s)
	h := handlers.NewHandler(srvs)
	r := httprouter.NewRouter(h)
	// запускаем сервер
	log.Println("accruals calculation service URL:", settings.ColorGreen, calcSys, settings.ColorReset)
	log.Println("starting http server on:", settings.ColorBlue, addr, settings.ColorReset)
	log.Fatal(http.ListenAndServe(addr, r))
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
		log.Println("eviroment variable RUN_ADDRESS is empty or has wrong value ", addr)
		addr = *addrFlag
	}
	// проверяем наличие переменной окружения, если ее нет или она не валидна, то используем значение из флага
	calcSys, ok = os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if !ok || !govalidator.IsURL(calcSys) || calcSys == "" {
		log.Println("eviroment variable ACCRUAL_SYSTEM_ADDRESS is empty or has wrong value ", calcSys)
		calcSys = *calcSysFlag
	}
	// проверяем наличие переменной окружения, если ее нет или она не валидна, то используем значение из флага
	dlink, ok = os.LookupEnv("DATABASE_URI")
	if !ok {
		log.Println("eviroment variable DATABASE_URI is not exist", dlink)
		dlink = *dlinkFlag
	}

	return dlink, calcSys, addr
}

// создание интерфейса хранилища
func newStrorageProvider(dlink string) (s services.StorageProvider) {
	// проверяем если переменная SQL url не пустая, логгируем
	if dlink == "" {
		log.Println("server may not properly start with "+settings.ColorRed+"database DSN empty", settings.ColorReset)
	}
	s = storage.NewSQLStorage(dlink)
	log.Println("server will start with data storage "+settings.ColorYellow+"in PostgreSQL:", dlink, settings.ColorReset)
	return s
}
