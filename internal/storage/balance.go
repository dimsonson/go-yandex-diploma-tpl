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
		// проверяем наличие сресдтв для списания, если недостаточно, возвращаем ошибку "insufficient funds"
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
			// логируем и возвращаем соответствующую ошибку "new order number already exist"
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
