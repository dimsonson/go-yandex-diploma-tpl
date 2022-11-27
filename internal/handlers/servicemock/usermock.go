package servicemock

import (
	"context"
	"errors"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"
)

type User struct {
}

func (msrv *User) Create(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	switch {
	case dc.Login == "dimma" && dc.Password == "12345":
		return nil
	case dc.Login == "dimma2login":
		return errors.New("login exist")
	default:
		return errors.New("something wrong woth server")
	}
}

func (msrv *User) CheckAuthorization(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	switch {
	case dc.Login == "dimma" && dc.Password == "12345":
		return nil
	case dc.Login == "dimma3login" && dc.Password == "12345":
		return errors.New("login or password not exist")
	default:
		return errors.New("something wrong woth server")
	}
}
