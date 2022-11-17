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
	 login    text NOT NULL UNIQUE,
	 password text NOT NULL
	 );
	
	CREATE TABLE IF NOT EXISTS orders
	(
	 order_num   text NOT NULL UNIQUE,
	 login       text NOT NULL,
	 change_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
	 status      text NOT NULL DEFAULT 'REGISTERED',
	 accrual     decimal DEFAULT 0	 
	);
		
	CREATE TABLE IF NOT EXISTS balance
	(
	 login           text NOT NULL UNIQUE,
	 current_balance decimal NOT NULL,
	 total_withdrawn decimal NOT NULL
	);
		
	CREATE TABLE IF NOT EXISTS withdrawals
	(
	 new_order       text NOT NULL UNIQUE,
	 login           text NOT NULL,
	 "sum"             decimal NOT NULL,
	 withdrawal_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
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
	// объявляем транзакцию
	tx, err := ms.PostgreSQL.BeginTx(ctx, nil)
	if err != nil {
		log.Println("error StorageNewOrderUpdate tx.Begin : ", err)
		return err
	}
	defer tx.Rollback()
	{
		// создаем текст запроса
		q := `INSERT INTO users VALUES ($1, $2)`
		// записываем в хранилице login, passwHex
		_, err = tx.Exec(q, login, passwHex)
		// если login есть в хранилище, возвращаем соответствующую ошибку
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
			err = errors.New("login exist")
			log.Println("insert 1st instruction of transaction StorageCreateNewUser SQL UniqueViolation error :", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Println("unable StorageCreateNewUser to rollback:", rollbackErr)
			}
			return err
		case err != nil && pgErr != nil && pgErr.Code != pgerrcode.UniqueViolation:
			log.Println("insert 1st instruction of transaction StorageCreateNewUser SQL request error :", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Println("unable StorageCreateNewUser to rollback:", rollbackErr)
			}
			return err
		default:
		}
	}
	{
		// создаем текст запроса
		q := `INSERT INTO balance VALUES ($1, 0, 0);`
		// записываем в хранилице login, passwHex
		_, err = tx.Exec(q, login)
		// если login есть в хранилище, возвращаем соответствующую ошибку
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
			err = errors.New("login exist")
			log.Println("insert 2st instruction of transaction StorageCreateNewUser SQL UniqueViolation error :", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Println("unable StorageCreateNewUser to rollback:", rollbackErr)
			}
			return err
		case err != nil && pgErr != nil && pgErr.Code != pgerrcode.UniqueViolation:
			log.Println("insert 2st instruction of transaction StorageCreateNewUser SQL request error :", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Println("unable StorageCreateNewUser to rollback:", rollbackErr)
			}
			return err
		default:
		}
	}
	// сохраняем изменения
	if err := tx.Commit(); err != nil {
		log.Println("error StorageNewOrderUpdate tx.Commit : ", err)
	}
	return err
}

// проверка наличия нового пользователя в хранилище - авторизация
func (ms *StorageSQL) StorageAuthorizationCheck(ctx context.Context, login string, passwHex string) (err error) {
	var exist int
	// создаем текст запроса
	q := `SELECT 1 FROM users WHERE login = $1 AND password = $2`
	// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную 
	err = ms.PostgreSQL.QueryRowContext(ctx, q, login, passwHex).Scan(&exist)
	if err != nil {
		log.Println("select StorageAuthorizationCheck SQL request scan error:", err)
	}
	if exist != 1 {
		err = errors.New("login or password not exist")
		return err
	}
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
	// объявляем транзакцию
	tx, err := ms.PostgreSQL.Begin()
	if err != nil {
		log.Println("error StorageNewOrderUpdate tx.Begin : ", err)
		return err
	}
	defer tx.Rollback()
	// создаем текст запроса обновление orders
	q := `UPDATE orders SET status = $3, accrual = $4 
	WHERE login = $1 AND order_num = $2 AND status != $3`
	// записываем в хранилице login, passwHex
	ordersUpd, err := ms.PostgreSQL.ExecContext(ctx, q, login, dc.Order, dc.Status, dc.Accrual)
	ordersRows, _ := ordersUpd.RowsAffected()
	// логируем и возвращаем соответствующую ошибку
	if err != nil || ordersRows != 1 {
		err = errors.New("error or row not found for order udpate")
		log.Println("update SQL request StorageNewOrderUpdate error:", err)
	}
	if dc.Accrual.GreaterThan(decimal.NewFromInt(0)) {
		// получаем текущее значение баланса аккаунта
		var balanceCurrent float64
		// создаем текст запроса
		q := `SELECT current_balance FROM balance WHERE login = $1`
		// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную value
		err = ms.PostgreSQL.QueryRowContext(ctx, q, login).Scan(&balanceCurrent)
		if err != nil {
			log.Println("select StorageNewOrderUpdate SQL request scan error:", err)
			return err
		}
		// добавляем значение начисления к балансу
		log.Println("balance before:", dc.Accrual)
		dc.Accrual.Add(decimal.NewFromFloat(balanceCurrent))
		log.Println("balance after:", dc.Accrual)
		// создаем текст запроса обновление balance
		q = `UPDATE balance SET current_balance = $2 WHERE login = $1`
		// записываем в хранилице login, passwHex
		balanceUpd, err := ms.PostgreSQL.ExecContext(ctx, q, login, dc.Accrual)
		balanceRows, _ := balanceUpd.RowsAffected()
		// если логируем и возвращаем соответствующую ошибку
		if err != nil && balanceRows != 1 {
			err = errors.New("error or row not found for balance udpate")
			log.Println("update SQL request StorageNewOrderUpdate error:", err)
		}
	}
	// сохраняем изменения
	if err := tx.Commit(); err != nil {
		log.Println("error StorageNewOrderUpdate tx.Commit : ", err)
	}
	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (ms *StorageSQL) StorageGetOrdersList(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	// создаем текст запроса
	q := `SELECT order_num, status, accrual, change_time FROM orders WHERE login = $1 ORDER BY change_time`
	// делаем запрос в SQL, получаем строку и пишем результат запроса в пременные
	rows, err := ms.PostgreSQL.QueryContext(ctx, q, login)
	if err != nil {
		log.Println("select StorageGetOrdersList SQL reqest error :", err)
		return ec, err
	}
	defer rows.Close()
	s := models.OrdersList{}
	// пишем результат запроса (итерирование по полученному набору строк) в структуру
	for rows.Next() {
		err = rows.Scan(&s.Number, &s.Status, &s.Accrual, &s.UploadedAt)
		if err != nil {
			log.Println("row by row scan StorageGetOrdersList error :", err)
			return ec, err
		}
		ec = append(ec, s)
	}
	// проверяем итерации на ошибки
	err = rows.Err()
	if err != nil {
		log.Println("request StorageGetOrdersList iteration scan error:", err)
		return ec, err
	}
	// проверяем наличие записей
	if len(ec) == 0 {
		err = errors.New("no orders for this login")
	}
	return ec, err
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (ms *StorageSQL) StorageGetUserBalance(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	// создаем текст запроса
	q := `SELECT current_balance, total_withdrawn FROM balance WHERE login = $1`
	// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную 
	err = ms.PostgreSQL.QueryRowContext(ctx, q, login).Scan(&ec.Current, &ec.Withdrawn)
	if err != nil {
		log.Println("select StorageAuthorizationCheck SQL request scan error:", err)
	}
	
	fmt.Println("StorageGetUserBalance login", login)

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
