package app

import (
	"context"
	"testing"

	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mokStore struct {
	getAddrFunc func(string) (string, error)
	setFunc     func(string, string) error
}

func (m mokStore) GetAddr(_ context.Context, s string) (string, error) { return m.getAddrFunc(s) }
func (m mokStore) Set(_ context.Context, s1 string, s2 string) error   { return m.setFunc(s1, s2) }
func (mokStore) Ping(context.Context) error                            { return nil }
func (mokStore) Close() error                                          { return nil }

// простой сокращатель
type mokShorter func(string) string

func (m mokShorter) Short(s string) string { return m(s) }

// Тесты

func newTestApp(t *testing.T) *Application {
	st := mokStore{
		setFunc: func(s1, s2 string) error { return nil },
		getAddrFunc: func(short string) (addr string, err error) {
			addr = short
			if short == "" {
				err = storage.ErrAddressNotFound
			}
			return
		},
	}
	sh := mokShorter(func(addr string) string { return addr })

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
