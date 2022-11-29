package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"

	//"github.com/dimsonson/go-yandex-diploma-tpl/internal/storage"
	"github.com/stretchr/testify/assert"
)

func NewPostgreProvider(db *sql.DB) *StorageSQL {
	return &StorageSQL{PostgreSQL: db}
}

func TestTodoItemPostgres_CheckAuthorization(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	fmt.Println(mock)

	r := NewPostgreProvider(db)
	//fmt.Println(r)

	type args struct {
		login    string
		passwHex string
	}
	type mockBehavior func(args args, err error)

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
				passwHex: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5", // `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImRpbW1hIn0.2RnFWU77EAKk9kQoxESrkiW8wjp22vWU-BrFvT3Z7Ro`,
			},
			want: nil,
			mock: func(args args, err error) {
				//mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"password"}).AddRow(args.passwHex)
				mock.ExpectQuery(`SELECT (.+) FROM users WHERE login = $1`).
					WithArgs(args.login). //, args.passwHex).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		/* 	{
			name: "Empty Fields",
			input: args{
				listId: 1,
				item: todo.TodoItem{
					Title:       "",
					Description: "description",
				},
			},
			mock: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id).RowError(0, errors.New("insert error"))
				mock.ExpectQuery("INSERT INTO todo_items").
					WithArgs(args.item.Title, args.item.Description).WillReturnRows(rows)

				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Failed 2nd Insert",
			input: args{
				listId: 1,
				item: todo.TodoItem{
					Title:       "title",
					Description: "description",
				},
			},
			mock: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO todo_items").
					WithArgs(args.item.Title, args.item.Description).WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO lists_items").WithArgs(args.listId, id).
					WillReturnError(errors.New("insert error"))

				mock.ExpectRollback()
			},
			wantErr: true,
		}, */
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input, tt.want)
			// создаем контекст
			ctx, cancel := context.WithTimeout(context.Background(), settings.StorageTimeout)
			// освобождаем ресурс
			defer cancel()
			err := r.CheckAuthorization(ctx, tt.input.login, tt.input.passwHex)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, err)
			}
			//	if err := mock.ExpectationsWereMet(); err != nil {
			//t.Errorf("there were unfulfilled expectations: %s", err)
			//}
		})
	}

}
