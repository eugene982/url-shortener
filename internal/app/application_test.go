package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewApplication(t *testing.T) {
	a := NewApplication(nil, nil)
	assert.IsType(t, &Application{}, a)
}

// простое хранилище
type mokStore struct {
	getAddrFunc  func(string) (string, bool)
	getShortFunc func(string) (string, bool)
	setFunc      func(string, string) bool
}

func (m mokStore) GetAddr(s string) (string, bool)  { return m.getAddrFunc(s) }
func (m mokStore) GetShort(s string) (string, bool) { return m.getShortFunc(s) }
func (m mokStore) Set(s1 string, s2 string) bool    { return m.setFunc(s1, s2) }

// простой сокращатель
type mokShorter func(string) string

func (m mokShorter) Short(s string) string { return m(s) }
