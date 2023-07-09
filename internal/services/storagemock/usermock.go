// пакет заглушек для получения ожидаемых ответов слоя storage при тестировании сервисов
package storagemock

import (
	"context"
	"errors"
)

type User struct {
}

func (mst *User) Create(ctx context.Context, login string, passwH string) (err error) {
	if login == "dimma" {
		return nil
	}
	err = errors.New("something wrong with server")
	return err

}

func (mst *User) CheckAuthorization(ctx context.Context, login string, passwHex string) (err error) {
	if login == "dimma" {
		return nil
	}
	err = errors.New("something wrong with server")
	return err
	
}
