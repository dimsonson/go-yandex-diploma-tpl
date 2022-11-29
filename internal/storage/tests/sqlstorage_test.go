package sqlstorage_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"

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
			name: "Ok",
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
