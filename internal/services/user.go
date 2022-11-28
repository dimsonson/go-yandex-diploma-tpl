package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/dimsonson/go-yandex-diploma-tpl/internal/models"

	"github.com/rs/zerolog/log"
)

// интерфейс методов хранилища для User
type UserStorageProvider interface {
	Create(ctx context.Context, login string, passwH string) (err error)
	CheckAuthorization(ctx context.Context, login string, passwHex string) (err error)
}

// структура конструктора бизнес логики User
type UserService struct {
	storage UserStorageProvider
}

// конструктор бизнес логики User
func NewUserService(uStorage UserStorageProvider) *UserService {
	return &UserService{
		uStorage,
	}
}

func (svc *UserService) Create(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	// сощдание хеш пароля для передачи в хранилище
	passwHex, err := ToHex(dc.Password)
	if err != nil {
		log.Printf("hex conversion in ServiceCreateNewUser error :%s", err)
		return err
	}
	// передача пары логин:пароль в хранилище
	err = svc.storage.Create(ctx, dc.Login, passwHex)
	return err
}

func (svc *UserService) CheckAuthorization(ctx context.Context, dc models.DecodeLoginPair) (err error) {
	// создание хеш пароля для передачи в хранилище
	passwHex, err := ToHex(dc.Password)
	if err != nil {
		log.Printf("hex conversion in ServiceCreateNewUser error :%s", err)
		return err
	}
	// передача пары логин:пароль в хранилище
	err = svc.storage.CheckAuthorization(ctx, dc.Login, passwHex)
	return err
}

// функция SHA.256 хеширования строки и кодирования хеша в строку
func ToHex(src string) (dst string, err error) {
	h := sha256.New()
	h.Write([]byte(src))
	tmpBytes := h.Sum(nil)
	dst = hex.EncodeToString(tmpBytes)
	return dst, err
}
