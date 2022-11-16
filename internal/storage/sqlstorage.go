package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
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
	q := `CREATE TABLE IF NOT EXISTS users
	(
	 login    text NOT NULL,
	 password text NOT NULL,
	 CONSTRAINT PK_1_users PRIMARY KEY ( login )
	);
	
	
	CREATE TABLE IF NOT EXISTS orders
	(
	 order_num   text NOT NULL,
	 login       text NOT NULL,
	 change_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
	 status      text NOT NULL DEFAULT 'REGISTERED',
	 accrual     decimal DEFAULT 0,
	 CONSTRAINT PK_1_orders PRIMARY KEY ( order_num ),
	 CONSTRAINT REF_FK_1_orders FOREIGN KEY ( login ) REFERENCES users ( login )
	);
	
	CREATE INDEX IF NOT EXISTS FK_1_orders ON orders
	(
	 login
	);
	
		
	CREATE TABLE IF NOT EXISTS balance
	(
	 login           text NOT NULL,
	 current         decimal NOT NULL,
	 total_withdrawn decimal NOT NULL,
	 CONSTRAINT PK_1_balance PRIMARY KEY ( login ),
	 CONSTRAINT REF_FK_4_balance FOREIGN KEY ( login ) REFERENCES users ( login )
	);
	
	CREATE INDEX IF NOT EXISTS FK_1_balance ON balance
	(
	 login
	);
	
	
	CREATE TABLE IF NOT EXISTS withdrawals
	(
	 new_order       text NOT NULL,
	 login           text NOT NULL,
	 "sum"             decimal NOT NULL,
	 withdrawal_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
	 CONSTRAINT PK_1_withdrawals PRIMARY KEY ( new_order ),
	 CONSTRAINT REF_FK_3_withdrawals FOREIGN KEY ( login ) REFERENCES users ( login )
	);
	
	CREATE INDEX IF NOT EXISTS FK_1_withdrawals ON withdrawals
	(
	 login
	);`

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
func (ms *StorageSQL) StorageCreateNewUser(ctx context.Context, login string, passwHex string) (err error) {
	// создаем текст запроса
	q := `INSERT INTO users VALUES ($1, $2)`
	// записываем в хранилице login, passwHex
	_, err = ms.PostgreSQL.ExecContext(ctx, q, login, passwHex)
	// если login есть в хранилище, возвращаем соответствующую ошибку
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		err = errors.New("login exist")
	}
	fmt.Println("StorageCreateNewUser login, passw", login, passwHex)
	return err
}

// проверка наличия нового пользователя в хранилище - авторизация
func (ms *StorageSQL) StorageAuthorizationCheck(ctx context.Context, login string, passwHex string) (err error) {
	var exist int
	// создаем текст запроса
	q := `SELECT 1 FROM users WHERE login = $1 AND password = $2`
	// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную value
	err = ms.PostgreSQL.QueryRowContext(ctx, q, login, passwHex).Scan(&exist)
	if err != nil {
		log.Println("select GetFromStorage SQL request scan error:", err)
		return err
	}
	fmt.Println(err)
	fmt.Println("StorageAuthorizationCheck login, passw", login, passwHex)
	return err
}

// сервис загрузки номера заказа для расчёта
func (ms *StorageSQL) StorageNewOrderLoad(ctx context.Context, login string, order_num string) (err error) {
	// создаем текст запроса
	q := `INSERT INTO orders (order_num, login) VALUES ($1, $2)`
	// записываем в хранилице login, passwHex
	_, err = ms.PostgreSQL.ExecContext(ctx, q, order_num, login)
	// если  есть в хранилище, возвращаем соответствующую ошибку
	var pgErr *pgconn.PgError
	// проверяем на UniqueViolation и получаем существующий логин для возврата ошибки в зависимости от того чей login
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		var existLogin string
		// создаем текст запроса
		q := `SELECT login FROM orders WHERE order_num = $1`
		// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную value
		err = ms.PostgreSQL.QueryRowContext(ctx, q, order_num).Scan(&existLogin)
		if err != nil {
			log.Println("select StorageNewOrderLoad SQL request scan error:", err)
			return err
		}
		if existLogin != login {
			err = errors.New("the same order number was loaded by another customer")
			return err
		}
		err = errors.New("order number from this login already exist")
		return err
	}
	if err != nil {
		return err
	}
	return err
}

// сервис обновление заказа для расчёта
func (ms *StorageSQL) StorageNewOrderUpdate(ctx context.Context, login string, dc models.OrderSatus) (err error) {
	// создаем текст запроса
	q := `UPDATE orders SET status = $3, accrual = $4 
	WHERE login = $1 AND order_num = $2 AND status != $3`
	// записываем в хранилице login, passwHex
	_, err = ms.PostgreSQL.ExecContext(ctx, q, login, dc.Order, dc.Status, dc.Accrual)
	// если логируем и возвращаем соответствующую ошибку
	if err != nil {
		log.Println("update SQL request StorageNewOrderUpdate error:", err)
	}
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
	fmt.Println("StorageGetUserBalance login", login)
	ec = models.LoginBalance{
		Current:   decimal.NewFromFloatWithExponent(500.505, -2),
		Withdrawn: decimal.NewFromFloatWithExponent(42, -2),
	}
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (ms *StorageSQL) StorageNewWithdrawal(login string, dc models.NewWithdrawal) (err error) {
	fmt.Println("StorageNewWithdrawal login, dc", login, dc)
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (ms *StorageSQL) StorageGetWithdrawalsList(login string) (ec models.WithdrawalsList, err error) {
	fmt.Println("StorageGetWithdrawalsList login", login)
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
