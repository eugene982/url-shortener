package storage

import (
	"context"
	"errors"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

// Хранитель ссылок
type Storage interface {
	Close() error
	Ping(context.Context) error
	GetAddr(ctx context.Context, short string) (addr string, err error)
	Set(ctx context.Context, addr string, short string) error
}
