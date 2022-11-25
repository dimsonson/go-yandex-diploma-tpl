package servicemock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

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
	fmt.Println(dc.Login, dc.Password)
	switch {
	case dc.Login == "dimma" && dc.Password == "12345":
		return nil
	case dc.Login == "dimma3login" && dc.Password == "12345":
		fmt.Println(dc.Login, dc.Password)
		return errors.New("login or password not exist")
	default:
		return errors.New("something wrong woth server")
	}
}

// функция SHA.256 хеширования строки и кодирования хеша в строку
func ToHex(src string) (dst string, err error) {
	h := sha256.New()
	h.Write([]byte(src))
	tmpBytes := h.Sum(nil)
	dst = hex.EncodeToString(tmpBytes)
	return dst, err
}
