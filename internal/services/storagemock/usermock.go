package storagemock

import (
	"context"
)

type User struct {
}

func (mst *User) Create(ctx context.Context, login string, passwH string) (err error) {
	return err

}

func (mst *User) CheckAuthorization(ctx context.Context, login string, passwHex string) (err error) {
	return nil
}
