package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/shopspring/decimal"
)

// структура хранилища
type StorageSQL struct {
	PostgreSQL *sql.DB
}

// конструктор нового хранилища PostgreSQL
func NewSQLStorage(p string) *StorageSQL {
	// создаем контекст и оснащаем его таймаутом
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, settings.StorageTimeout)
	defer cancel()
	// открываем базу данных
	db, err := sql.Open("pgx", p)
	if err != nil {
		log.Println("database opening error:", settings.ColorRed, err, settings.ColorReset)
	}
	// создаем текст запроса
	q := `CREATE TABLE IF NOT EXISTS sh_urls (
				"userid" TEXT,
				"short_url" TEXT NOT NULL UNIQUE,
				"long_url" TEXT NOT NULL UNIQUE,
				"deleted_url" BOOLEAN 
				)`
	// создаем таблицу в SQL базе, если не существует
	_, err = db.ExecContext(ctx, q)
	if err != nil {
		log.Println("request NewSQLStorage to sql db returned error:", settings.ColorRed, err, settings.ColorReset)
	}
	return &StorageSQL{
		PostgreSQL: db,
	}
}

// метод закрытия совединения с SQL базой
func (ms *StorageSQL) StorageConnectionClose() {
	ms.PostgreSQL.Close()
}

// добавление нового пользователя в хранилище
func (ms *StorageSQL) StorageCreateNewUser(login string, passw []byte) (err error) {
	fmt.Println("StorageCreateNewUser login, passw", login, string(passw))
	return err
}

// проверка наличия нового пользователя в хранилище - авторизация
func (ms *StorageSQL) StorageAuthorizationCheck(login string, passwHex string) (err error) {
	fmt.Println("StorageAuthorizationCheck login, passw", login, passwHex)
	return err
}

// сервис загрузки пользователем номера заказа для расчёта
func (ms *StorageSQL) StorageNewOrderLoad(login string, order_num string) (err error) {
	fmt.Println("StorageNewOrderLoad login, order_num ", login, order_num)
	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (ms *StorageSQL) StorageGetOrdersList(login string) (ec models.OrdersList, err error) {
	fmt.Println("StorageGetOrdersList login", login)
	ec = models.OrdersList{
		{
			Number:  "9278923470",
			Status:  "PROCESSED",
			Accrual: decimal.NewFromFloatWithExponent(500, -2),
			// UploadedAt: "2020-12-10T15:15:45+03:00",
		},
		{
			Number: "12345678903",
			Status: "PROCESSING",
			// UploadedAt: "2020-12-10T15:12:01+03:00",
		},
		{
			Number: "346436439",
			Status: "INVALID",
			// UploadedAt: "2020-12-09T16:09:53+03:00",
		},
	}
	return ec, err
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (ms *StorageSQL) StorageGetUserBalance(login string) (ec models.LoginBalance, err error) {
	fmt.Println("ServiceGetUserBalance login", login)
	ec = models.LoginBalance{
		Current:   decimal.NewFromFloatWithExponent(500.505, -2),
		Withdrawn: decimal.NewFromFloatWithExponent(42, -2),
	}
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (ms *StorageSQL) StorageNewWithdrawal(login string, dc models.NewWithdrawal) (err error) {
	fmt.Println("ServiceNewWithdrawal login, dc", login, dc)
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (ms *StorageSQL) StorageGetWithdrawalsList(login string) (ec models.WithdrawalsList, err error) {
	fmt.Println("ServiceGetWithdrawalsList login", login)
	ec = models.WithdrawalsList{
		{
			Order: "2377225624",
			Sum:   decimal.NewFromFloatWithExponent(500.0300, -2),
			//ProcessedAt: "2020-12-09T16:09:57+03:00",
		},
		{
			Order: "2377225625",
			Sum:   decimal.NewFromFloatWithExponent(800.5555, -2),
			//ProcessedAt: "2020-12-09T16:09:57+03:00",
		},
	}
	return ec, err
}
