package sqlstorage_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"
	"github.com/jackc/pgconn"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/storage"
	"github.com/stretchr/testify/assert"
)

func NewPostgreProvider(db *sql.DB) *storage.StorageSQL {
	return &storage.StorageSQL{PostgreSQL: db}
}

func TestStorage_CheckAuthorization(t *testing.T) {
	// создание заглушки
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	r := NewPostgreProvider(db)
	// принимаемые аргументы
	type args struct {
		login    string
		passwHex string
	}
	// тип поведения заглушки
	type mockBehavior func(args args, err error)
	// табличный тест, будет дополнен негатвными сценариями
	tests := []struct {
		name    string
		mock    mockBehavior
		input   args
		want    error
		wantErr bool
	}{
		{
			name: "Positive test - autorization",
			input: args{
				login:    "dimma",
				passwHex: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			},
			want: nil,
			mock: func(args args, err error) {
				rows := sqlmock.NewRows([]string{"password"}).AddRow(args.passwHex)
				mock.ExpectQuery(`SELECT (.+) FROM users WHERE login (.+)`).
					WithArgs(args.login).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Negative test - wrong passw or login",
			input: args{
				login:    "dimma",
				passwHex: "1",
			},
			want: errors.New("login or password not exist"),
			mock: func(args args, err error) {
				rows := sqlmock.NewRows([]string{"password"}).AddRow("2")
				mock.ExpectQuery(`SELECT (.+) FROM users WHERE login (.+)`).
					WithArgs(args.login).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "Negative test - sql DB down",
			input: args{
				login:    "dimma",
				passwHex: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			},
			want: errors.New("FATAL: terminating connection due to administrator command (SQLSTATE 57P01)"),
			mock: func(args args, err error) {
				mock.ExpectQuery(`SELECT (.+) FROM users WHERE login (.+)`).
					WithArgs(args.login).
					WillReturnError(errors.New("FATAL: terminating connection due to administrator command (SQLSTATE 57P01)"))

			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input, tt.want)
			// создаем контекст
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			// запуск метода запроса
			err := r.CheckAuthorization(ctx, tt.input.login, tt.input.passwHex)
			// проверки
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestStorage_Create(t *testing.T) {
	// создание заглушки
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	r := NewPostgreProvider(db)
	// принимаемые аргументы
	type args struct {
		login    string
		passwHex string
	}
	// тип поведения заглушки
	type mockBehavior func(args args, err error)
	// ошибка нарушение уникальности значений при вставке
	duplicateErr := &pgconn.PgError{
		Severity:         "ERROR",
		Code:             "23505",
		Message:          "duplicate key value violates unique constraint \"pk_1_users\"",
		Detail:           "Key (login)=(dimma) already exists.",
		Hint:             "",
		Position:         0,
		InternalPosition: 0,
		InternalQuery:    "",
		Where:            "",
		SchemaName:       "public",
		TableName:        "users",
		ColumnName:       "",
		DataTypeName:     "",
		ConstraintName:   "pk_1_users",
		File:             "nbtinsert.c",
		Line:             670,
		Routine:          "_bt_check_unique",
	}

	// табличный тест
	tests := []struct {
		name    string
		mock    mockBehavior
		input   args
		want    error
		wantErr bool
	}{
		{
			name: "Positive test - create user",
			input: args{
				login:    "dimma",
				passwHex: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			},
			want: nil,
			mock: func(args args, err error) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users VALUES (.+)`).
					WithArgs(args.login, args.passwHex).WillReturnResult(sqlmock.NewResult(0, 1)).WillReturnError(nil)
				mock.ExpectExec(`INSERT INTO balance VALUES (.+)`).
					WithArgs(args.login).WillReturnResult(sqlmock.NewResult(0, 1)).WillReturnError(nil)
			},

			wantErr: false,
		},
		{
			name: "Negative test - create user - login exist - error at 1nd sql instruction",
			input: args{
				login:    "dimma2",
				passwHex: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			},
			want: errors.New("login exist"),
			mock: func(args args, err error) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users VALUES (.+)`).
					WithArgs(args.login, args.passwHex).WillReturnError(duplicateErr)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Negative test - create user - database down at 1st sql instruction",
			input: args{
				login:    "dimma3",
				passwHex: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			},
			want: errors.New("FATAL: terminating connection due to administrator command (SQLSTATE 57P01)"),
			mock: func(args args, err error) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users VALUES (.+)`).
					WithArgs(args.login, args.passwHex).WillReturnError(errors.New(`FATAL: terminating connection due to administrator command (SQLSTATE 57P01)`))
				mock.ExpectRollback()
			},
			wantErr: true,
		},

		{
			name: "Negative test - database down at 2nd sql instruction",
			input: args{
				login:    "dimma",
				passwHex: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			},
			want: errors.New("FATAL: terminating connection due to administrator command (SQLSTATE 57P01)"),
			mock: func(args args, err error) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users VALUES (.+)`).
					WithArgs(args.login, args.passwHex).WillReturnResult(sqlmock.NewResult(0, 1)).WillReturnError(nil)
				mock.ExpectExec(`INSERT INTO balance VALUES (.+)`).
					WithArgs(args.login).WillReturnError(errors.New(`FATAL: terminating connection due to administrator command (SQLSTATE 57P01)`))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Negative test - create user - login exist - err at 2nd sql instruction",
			input: args{
				login:    "dimma",
				passwHex: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			},
			want: errors.New("login exist"),
			mock: func(args args, err error) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO users VALUES (.+)`).
					WithArgs(args.login, args.passwHex).WillReturnResult(sqlmock.NewResult(0, 1)).WillReturnError(nil)
				mock.ExpectExec(`INSERT INTO balance VALUES (.+)`).
					WithArgs(args.login).WillReturnError(duplicateErr)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input, tt.want)
			// создаем контекст
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			// запуск метода запроса
			err := r.Create(ctx, tt.input.login, tt.input.passwHex)
			// проверки
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.want, err)
			} else {
				assert.NoError(t, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
