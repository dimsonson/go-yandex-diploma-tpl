package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/rs/zerolog/log"
)

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
