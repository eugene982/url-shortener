// Тестирование хранилища

package storage

import (
	"testing"
)

func TestGetAddr(t *testing.T) {

	var cases = []struct {
		addr  string
		short string
	}{
		{"", ""},
		{"ya.ru", ""},
		{"", "t1"},
		{"ya.ru", "t1"},
		{"http://ya.ru", "t1"},
		{"https://ya.ru", "t1"},
		{"https://yandex.ru", "t1"},
		{"http://ya.ru", "t1"},
		{"ya.ru", "t2"},
	}

	storage := NewMemstore()

	for _, c := range cases {

		storage.Set(c.addr, c.short)

		if get, ok := storage.GetAddr(c.short); !ok || c.addr != get {
			t.Errorf("error get addr: short: %s, get %s, want %s",
				c.short, get, c.addr)
			continue
		}
	}
}
