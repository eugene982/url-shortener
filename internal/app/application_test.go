package app

import (
	"context"
	"testing"

	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mokStore struct {
	getAddrFunc     func(string) (string, error)
	updFunc         func(d ...model.StoreData) error
	getUserURLsFunc func() ([]model.StoreData, error)
}

func (m mokStore) GetAddr(_ context.Context, s string) (string, error) { return m.getAddrFunc(s) }
func (m mokStore) Set(_ context.Context, d model.StoreData) error {
	return m.updFunc(d)
}
func (m mokStore) GetUserURLs(_ context.Context, userID int64) ([]model.StoreData, error) {
	return m.getUserURLsFunc()
}

func (m mokStore) Update(_ context.Context, ls []model.StoreData) error { return m.updFunc(ls...) }
func (mokStore) Ping(context.Context) error                             { return nil }
func (mokStore) Close() error                                           { return nil }

// простой сокращатель
type mokShorter func(string) (string, error)

func (m mokShorter) Short(s string) (string, error) { return m(s) }

// Тесты

func newTestApp(t *testing.T) *Application {
	st := mokStore{
		updFunc: func(_ ...model.StoreData) error { return nil },
		getAddrFunc: func(short string) (addr string, err error) {
			addr = short
			if short == "" {
				err = storage.ErrAddressNotFound
			}
			return
		},
		getUserURLsFunc: func() ([]model.StoreData, error) {
			return []model.StoreData{}, nil
		},
	}
	sh := mokShorter(func(addr string) (string, error) { return addr, nil })

	a, err := NewApplication(sh, st, "")
	require.NotNil(t, a)
	require.NoError(t, err)

	return a
}

func TestNewApplication(t *testing.T) {
	a := newTestApp(t)
	assert.IsType(t, &Application{}, a)

	err := a.Close()
	require.NoError(t, err)
}
