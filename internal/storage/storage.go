package storage

import (
	"context"
	"errors"

	"github.com/eugene982/url-shortener/internal/model"
)

var (
	// ошибка возвращается если по указанному короткому адресу не наден полный адрес
	ErrAddressNotFound = errors.New("address not found")

	// ошибка возвращается при наличи уже сохраненного адреса
	ErrAddressConflict = errors.New("address conflict")
)

// Storage интрефейс хранилища ссылок пользователей
type Storage interface {
	Close() error
	Ping(context.Context) error
	GetAddr(ctx context.Context, short string) (data model.StoreData, err error)
	Set(ctx context.Context, data model.StoreData) error
	Update(ctx context.Context, list []model.StoreData) error
	GetUserURLs(ctx context.Context, userID string) ([]model.StoreData, error)
	DeleteShort(ctx context.Context, shortURLs []string) error
}
