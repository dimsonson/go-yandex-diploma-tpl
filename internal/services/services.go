package services

import "github.com/dimsonson/go-yandex-diploma-tpl/internal/settings"

// интерфейс методов хранилища
type StorageProvider interface {
	StorageConnectionClose()
}

// структура конструктора бизнес логики
type Services struct {
	storage StorageProvider
}

// конструктор бизнес логики
func NewService(s StorageProvider) *Services {
	return &Services{
		s,
	}
}

func (sr *Services) ServiceCreateNewUser(dc settings.DecodeLoginPair) (err error) {
	return err
}

func (sr *Services) ServiceAuthorizationCheck(dc settings.DecodeLoginPair) (err error) {
	return err
}

func (sr *Services) ServiceNewOrderLoad(login string, order_num string) (err error) {
	return err
}
