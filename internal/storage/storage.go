package storage

import (
	"context"
	"errors"

	"github.com/eugene982/url-shortener/internal/model"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

// Хранитель ссылок
type Storage interface {
	Close() error
	Ping(context.Context) error
	GetAddr(ctx context.Context, short string) (addr string, err error)
	Set(ctx context.Context, data ...model.StoreData) error
}
