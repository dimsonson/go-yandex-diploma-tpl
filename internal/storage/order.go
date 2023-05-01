package storage

import (
	"context"
	"errors"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

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
