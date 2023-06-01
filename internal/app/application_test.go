package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mokStore struct {
	getAddrFunc func(string) (string, bool)
	setFunc     func(string, string) error
}

func (m mokStore) GetAddr(s string) (string, bool) { return m.getAddrFunc(s) }
func (m mokStore) Set(s1 string, s2 string) error  { return m.setFunc(s1, s2) }
func (mokStore) Close() error                      { return nil }

// простой сокращатель
type mokShorter func(string) string

func (m mokShorter) Short(s string) string { return m(s) }

// Тесты

func newTestApp(t *testing.T) *Application {
	st := mokStore{
		getAddrFunc: func(addr string) (string, bool) { return addr, addr != "" },
		setFunc:     func(s1, s2 string) error { return nil },
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
