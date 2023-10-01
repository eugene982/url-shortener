package app

import (
	"context"
	"testing"
	"time"

	"github.com/eugene982/url-shortener/internal/config"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mokStore struct {
	getAddrFunc     func(string) (model.StoreData, error)
	updFunc         func(d ...model.StoreData) error
	getUserURLsFunc func() ([]model.StoreData, error)
	getStats        func() (int, int, error)
}

func (m mokStore) GetAddr(_ context.Context, s string) (model.StoreData, error) {
	return m.getAddrFunc(s)
}
func (m mokStore) Set(_ context.Context, d model.StoreData) error {
	return m.updFunc(d)
}
func (m mokStore) GetUserURLs(_ context.Context, userID string) ([]model.StoreData, error) {
	return m.getUserURLsFunc()
}
func (m mokStore) DeleteShort(ctx context.Context, shortURLs []string) error  { return nil }
func (m mokStore) Stats(ctx context.Context) (URLs int, users int, err error) { return m.getStats() }
func (m mokStore) Update(_ context.Context, ls []model.StoreData) error       { return m.updFunc(ls...) }
func (mokStore) Ping(context.Context) error                                   { return nil }
func (mokStore) Close() error                                                 { return nil }

// простой сокращатель
type mokShorter func(string) (string, error)

func (m mokShorter) Short(s string) (string, error) { return m(s) }

// Тесты

func newTestApp(t *testing.T) *Application {
	conf := config.Configuration{
		ServAddr: "localhost:8080",
		ProfAddr: "localhost:8081",
	}

	a, err := New(conf)
	require.NotNil(t, a)
	require.NoError(t, err)

	a.shortener = mokShorter(func(addr string) (string, error) { return addr, nil })
	a.store = mokStore{
		updFunc: func(_ ...model.StoreData) error { return nil },
		getAddrFunc: func(short string) (data model.StoreData, err error) {
			if short == "" {
				err = storage.ErrAddressNotFound
			}
			data.OriginalURL = short
			data.ShortURL = short
			return
		},
		getUserURLsFunc: func() ([]model.StoreData, error) {
			return []model.StoreData{}, nil
		},
	}

	return a
}

func TestNewApplication(t *testing.T) {
	a := newTestApp(t)
	assert.IsType(t, &Application{}, a)

	go func() {
		err := a.Start()
		require.NoError(t, err)
	}()
	go a.startDeletionShortUrls()
	a.DeleteUserShortAsync("user", []string{"ya.ru"})

	time.Sleep(time.Second)

	err := a.Stop()
	require.NoError(t, err)
}

func TestNewApplicationHTTPS(t *testing.T) {
	conf := config.Configuration{
		EnableHTTPS: true,
	}
	a, err := New(conf)
	require.NoError(t, err)

	err = a.Start()
	require.Error(t, err)
}

func TestNewApplicationDSN(t *testing.T) {
	conf := config.Configuration{
		DatabaseDSN: "postgres://...",
	}
	_, err := New(conf)
	require.Error(t, err)
}
