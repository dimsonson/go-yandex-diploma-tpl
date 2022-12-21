// пакет хранилища PostgreSQL
package storage

import (
	"context"
	"database/sql"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog/log"
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
	// проверяем соединение с postgres
	err = db.PingContext(ctx)
	if err != nil {
		log.Printf("database connection is not alive: %s%s%s", settings.ColorRed, err, settings.ColorReset)
		return nil
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
