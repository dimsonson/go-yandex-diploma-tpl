package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog/log"
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
		log.Printf("database opening error: %s%s%s", settings.ColorRed, err, settings.ColorReset)
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
	 status      text NOT NULL DEFAULT 'NEW',
	 accrual     decimal DEFAULT 0,	 
	 CONSTRAINT PK_1_orders PRIMARY KEY ( order_num ),
	 CONSTRAINT REF_FK_1_orders FOREIGN KEY ( login ) REFERENCES users ( login )
	);
		
	CREATE TABLE IF NOT EXISTS balance
	(
	 login           text NOT NULL UNIQUE,
	 current_balance decimal NOT NULL,
	 total_withdrawn decimal NOT NULL,
	 CONSTRAINT PK_1_balance PRIMARY KEY ( login ),
	 CONSTRAINT REF_FK_4_balance FOREIGN KEY ( login ) REFERENCES users ( login )
	);
		
	CREATE TABLE IF NOT EXISTS withdrawals
	(
	 new_order       text NOT NULL UNIQUE,
	 login           text NOT NULL,
	 "sum"             decimal NOT NULL,
	 withdrawal_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
	 CONSTRAINT PK_1_withdrawals PRIMARY KEY ( new_order ),
	 CONSTRAINT REF_FK_3_withdrawals FOREIGN KEY ( login ) REFERENCES users ( login )
	);`

	// создаем таблицу в SQL базе, если не существует
	_, err = db.ExecContext(ctx, q)
	if err != nil {
		log.Printf("request NewSQLStorage to sql db returned error: %s%s%s", settings.ColorRed, err, settings.ColorReset)
	}
	return &StorageSQL{
		PostgreSQL: db,
	}
}

// метод закрытия совединения с SQL базой
func (ms *StorageSQL) ConnectionClose() {
	ms.PostgreSQL.Close()
}

// добавление нового пользователя в хранилище, запись в две таблицы в транзакции
func (ms *StorageSQL) Create(ctx context.Context, login string, passwHex string) (err error) {
	// объявляем транзакцию
	tx, err := ms.PostgreSQL.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("error StorageNewOrderUpdate tx.Begin : %s", err)
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
			log.Printf("insert 1st instruction of transaction StorageCreateNewUser SQL UniqueViolation error : %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
			}
			return err
		case err != nil:
			log.Printf("insert 1st instruction of transaction StorageCreateNewUser SQL request error : %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
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
			log.Printf("insert 2nd instruction of transaction StorageCreateNewUser SQL UniqueViolation error : %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
			}
			return err
		case err != nil:
			log.Printf("insert 2nd instruction of transaction StorageCreateNewUser SQL request error : %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
			}
			return err
		default:
		}
	}
	// сохраняем изменения
	if err := tx.Commit(); err != nil {
		log.Printf("error StorageNewOrderUpdate tx.Commit : %s", err)
	}
	return err
}

// проверка наличия нового пользователя в хранилище - авторизация
func (ms *StorageSQL) CheckAuthorization(ctx context.Context, login string, passwHex string) (err error) {
	var passwDB string
	// создаем текст запроса
	q := `SELECT password FROM users WHERE login = $1`
	// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную
	err = ms.PostgreSQL.QueryRowContext(ctx, q, login).Scan(&passwDB)
	if err != nil {
		log.Printf("select StorageAuthorizationCheck SQL request scan error: %s", err)
	}
	if passwDB != passwHex {
		err = errors.New("login or password not exist")
		log.Printf("select StorageAuthorizationCheck SQL: %s", err)
		return err
	}
	return err
}

// сервис загрузки номера заказа для расчёта без обноления статуса
func (ms *StorageSQL) Load(ctx context.Context, login string, orderNum string) (err error) {
	// создаем текст запроса, часть значений дефолтные в DB Postgre (см конструктор базы)
	q := `INSERT INTO orders (order_num, login) VALUES ($1, $2)`
	// записываем в хранилице orderNum, login
	_, err = ms.PostgreSQL.ExecContext(ctx, q, orderNum, login)
	// если нет ошибки, возвращаем nil
	if err == nil {
		return err
	}
	// переменная ошибки sql
	var pgErr *pgconn.PgError
	// проверяем на UniqueViolation и получаем существующий логин для возврата ошибки в зависимости от того чей login
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		var existLogin string
		// создаем текст запроса
		q := `SELECT login FROM orders WHERE order_num = $1`
		// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную value
		err = ms.PostgreSQL.QueryRowContext(ctx, q, orderNum).Scan(&existLogin)
		if err != nil {
			log.Printf("select StorageNewOrderLoad SQL request scan error: %s", err)
			return err
		}
		if existLogin != login {
			err = errors.New("the same order number was loaded by another customer")
			log.Printf("select StorageNewOrderLoad SQL request: %s", err)
			return err
		}
		err = errors.New("order number from this login already exist")
		log.Printf("select StorageNewOrderLoad SQL request : %s", err)
		return err
	}
	log.Printf("insert StorageNewOrderLoad error : %s", err)
	return err
}

// сервис обновление статуса и начислений заказа для расчёта
func (ms *StorageSQL) Update(ctx context.Context, login string, dc models.OrderSatus) (err error) {
	// объявляем транзакцию
	tx, err := ms.PostgreSQL.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("error StorageNewOrderUpdate tx.Begin : %s", err)
		return err
	}
	defer tx.Rollback()
	{
		// создаем текст запроса обновление orders
		q := `UPDATE orders SET status = $3, accrual = $4 WHERE login = $1 AND order_num = $2 AND status != $3`
		// записываем в хранилице поля из структуры и аргумента
		_, err := ms.PostgreSQL.Exec(q, login, dc.Order, dc.Status, dc.Accrual)
		// логируем и возвращаем соответствующую ошибку
		if err != nil {
			log.Printf("update SQL request StorageNewOrderUpdate error: %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
			}
			return err
		}
	}
	{
		//если сумма начистения в обновлении больше 0, то добавлеям сумму начисления к балансу
		if dc.Accrual.GreaterThan(decimal.NewFromInt(0)) {
			// получаем текущее значение баланса аккаунта
			var balanceCurrent decimal.Decimal
			// создаем текст запроса
			q := `SELECT current_balance FROM balance WHERE login = $1`
			// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную
			err = ms.PostgreSQL.QueryRow(q, login).Scan(&balanceCurrent)
			if err != nil {
				log.Printf("select StorageNewOrderUpdate SQL request scan error: %s", err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
				}
				return err
			}
			// добавляем значение начисления к балансу
			log.Printf("balance before: %s", balanceCurrent)
			balanceCurrent = dc.Accrual.Add(balanceCurrent)
			log.Printf("balance after: %s", balanceCurrent)
			// создаем текст запроса обновление balance
			q = `UPDATE balance SET current_balance = $2 WHERE login = $1`
			// записываем в хранилице
			_, err := ms.PostgreSQL.Exec(q, login, dc.Accrual)
			// если не ок логируем и возвращаем соответствующую ошибку
			if err != nil {
				err = errors.New("error for balance udpate")
				log.Printf("update SQL request StorageNewOrderUpdate error: %s", err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
				}
				return err
			}
		}
	}
	// сохраняем изменения
	if err := tx.Commit(); err != nil {
		log.Printf("error StorageNewOrderUpdate tx.Commit %s: ", err)
	}
	return err
}

// сервис получения списка размещенных пользователем заказов, сортировка выдачи по времени загрузки
func (ms *StorageSQL) List(ctx context.Context, login string) (ec []models.OrdersList, err error) {
	// создаем текст запроса
	q := `SELECT order_num, status, accrual, change_time FROM orders WHERE login = $1 ORDER BY change_time`
	// делаем запрос в SQL, получаем строку и пишем результат запроса в пременные
	rows, err := ms.PostgreSQL.QueryContext(ctx, q, login)
	if err != nil {
		log.Printf("select StorageGetOrdersList SQL reqest error %s:", err)
		return ec, err
	}
	defer rows.Close()
	s := models.OrdersList{}
	// пишем результат запроса (итерирование по полученному набору строк) в структуру
	for rows.Next() {
		err = rows.Scan(&s.Number, &s.Status, &s.Accrual, &s.UploadedAt)
		if err != nil {
			log.Printf("row by row scan StorageGetOrdersList error : %s", err)
			return ec, err
		}
		ec = append(ec, s)
	}
	// проверяем итерации на ошибки
	err = rows.Err()
	if err != nil {
		log.Printf("request StorageGetOrdersList iteration scan error: %s", err)
		return ec, err
	}
	// проверяем наличие записей
	if len(ec) == 0 {
		log.Printf("request StorageGetOrdersList len == 0: %s", err)
		err = errors.New("no orders for this login")
	}
	return ec, err
}

// сервис получение текущего баланса счёта баллов лояльности пользователя
func (ms *StorageSQL) Status(ctx context.Context, login string) (ec models.LoginBalance, err error) {
	// создаем текст запроса
	q := `SELECT current_balance, total_withdrawn FROM balance WHERE login = $1`
	// делаем запрос в SQL, получаем строку и пишем результат запроса в пременную
	err = ms.PostgreSQL.QueryRowContext(ctx, q, login).Scan(&ec.Current, &ec.Withdrawn)
	if err != nil {
		log.Printf("select StorageAuthorizationCheck SQL request scan error: %s", err)
	}
	return ec, err
}

// сервис списание баллов с накопительного счёта в счёт оплаты нового заказа
func (ms *StorageSQL) NewWithdrawal(ctx context.Context, login string, dc models.NewWithdrawal) (err error) {
	// объявляем транзакцию
	tx, err := ms.PostgreSQL.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("error StorageNewOrderUpdate tx.Begin : %s", err)
		return err
	}
	defer tx.Rollback()
	{
		// уменьшаем остаток баланса на сумму списания и увеличиваем общую сумму списаний на эту же смумму
		// получаем текущее значение баланса аккаунта и общую сумму списаний
		var balanceCurrent decimal.Decimal
		var balanceWithdrawls decimal.Decimal
		// создаем текст запроса
		q := `SELECT current_balance, total_withdrawn FROM balance WHERE login = $1`
		// делаем запрос в SQL, получаем строку и пишем результат запроса в пременные
		err = ms.PostgreSQL.QueryRow(q, login).Scan(&balanceCurrent, &balanceWithdrawls)
		if err != nil {
			log.Printf("select StorageNewOrderUpdate SQL request scan error: %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
			}
			return err
		}
		// проверяем наличие сресдтв для списания
		if dc.Sum.GreaterThan(balanceCurrent) {
			err = errors.New("insufficient funds")
			log.Printf("error StorageNewWithdrawal : %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
			}
			return err
		}
		// добавляем значение списания к балансу списаний
		log.Printf("balanceWithdrawls before: %s", balanceWithdrawls)
		balanceWithdrawls = balanceWithdrawls.Add(dc.Sum)
		log.Printf("balanceWithdrawls after: %s", balanceWithdrawls)

		// вычитаем значение списания из баланса счета
		log.Printf("balanceCurrent before: %s", balanceCurrent)
		balanceCurrent = balanceCurrent.Sub(dc.Sum)
		log.Printf("balanceCurrent after: %s", balanceCurrent)

		// создаем текст запроса обновление balance
		q = `UPDATE balance SET current_balance = $2, total_withdrawn =$3 WHERE login = $1`
		// записываем в хранилице
		_, err := ms.PostgreSQL.Exec(q, login, balanceCurrent, balanceWithdrawls)
		if err != nil {
			log.Printf("update StorageNewOrderUpdate SQL request error: %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
			}
			return err
		}
	}
	{
		{ // создаем текст запроса обновление withdrawals
			q := `INSERT INTO withdrawals (new_order, login, "sum") VALUES ($1, $2, $3)`
			// записываем в хранилице login, passwHex
			_, err := ms.PostgreSQL.Exec(q, dc.Order, login, dc.Sum)
			// логируем и возвращаем соответствующую ошибку
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				log.Printf("error StorageNewWithdrawal : %s", err)
				err = errors.New("new order number already exist")
				return err
			}
			if err != nil {
				log.Printf("update SQL request StorageNewOrderUpdate error: %s", err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
				}
				return err
			}
		}
		// если не ок логируем и возвращаем соответствующую ошибку
		if err != nil {
			log.Printf("update SQL request StorageNewOrderUpdate error: %s", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("unable StorageCreateNewUser to rollback: %s", rollbackErr)
			}
			return err
		}
	}
	// сохраняем изменения
	if err := tx.Commit(); err != nil {
		log.Printf("error StorageNewOrderUpdate tx.Commit : %s", err)
	}
	return err
}

// сервис информации о всех выводах средств с накопительного счёта пользователем
func (ms *StorageSQL) WithdrawalsList(ctx context.Context, login string) (ec []models.WithdrawalsList, err error) {
	// создаем текст запроса
	q := `SELECT new_order, "sum", withdrawal_time FROM withdrawals WHERE login = $1 ORDER BY withdrawal_time`
	// делаем запрос в SQL, получаем строку и пишем результат запроса в пременные
	rows, err := ms.PostgreSQL.QueryContext(ctx, q, login)
	if err != nil {
		log.Printf("select StorageGetWithdrawalsList SQL reqest error : %s", err)
		return ec, err
	}
	defer rows.Close()
	s := models.WithdrawalsList{}
	// пишем результат запроса (итерирование по полученному набору строк) в структуру
	for rows.Next() {
		err = rows.Scan(&s.Order, &s.Sum, &s.ProcessedAt)
		if err != nil {
			log.Printf("row by row scan StorageGetWithdrawalsList error : %s", err)
			return ec, err
		}
		ec = append(ec, s)
	}
	// проверяем итерации на ошибки
	err = rows.Err()
	if err != nil {
		log.Printf("request StorageGetWithdrawalsList iteration scan error: %s", err)
		return ec, err
	}
	// проверяем наличие записей
	if len(ec) == 0 {
		err = errors.New("no records")
		log.Printf("request StorageGetWithdrawalsList len == 0: %s", err)
	}
	return ec, err
}
