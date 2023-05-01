package app

import "testing"

func TestSet(t *testing.T) {

	var cases = []struct {
		addr  string
		short string
		ok    bool
	}{
		{"", "", false},
		{"", "t0", false},
		{"ya.ru", "", false},
		{"ya.ru", "t1", true},
		{"http://ya.ru", "t2", true},
	}

	storage := NewMemstore()

	for _, c := range cases {
		ok := storage.Set(c.addr, c.short)
		if c.ok != ok {
			t.Errorf("error set addr: '%s' (%s) - get %v, want %v",
				c.addr, c.short, ok, c.ok)
			continue
		}
	}
}

func TestGetShort(t *testing.T) {

	var cases = []struct {
		addr  string
		short string
	}{
		{"ya.ru", "t1"},
		{"ya.ru", "t2"},
		{"https://ya.ru", "t3"},
		{"ya.ru", "t4"},
		{"ya.ru", "t5"},
	}

	storage := NewMemstore()

	for _, c := range cases {

		if !storage.Set(c.addr, c.short) {
			t.Errorf("error set addr: '%s' (%s)",
				c.addr, c.short)
		}

		if get, ok := storage.GetShort(c.addr); !ok || c.short != get {
			t.Errorf("error get short: addr %s, get %s, want %s",
				c.addr, get, c.short)
		}
	}
}

func TestGetAddr(t *testing.T) {

	var cases = []struct {
		addr  string
		short string
	}{
		{"ya.ru", "t1"},
		{"http://ya.ru", "t1"},
		{"https://ya.ru", "t1"},
		{"https://yandex.ru", "t1"},
		{"http://ya.ru", "t1"},
		{"ya.ru", "t2"},
	}

	storage := NewMemstore()

	for _, c := range cases {

		if !storage.Set(c.addr, c.short) {
			t.Errorf("error set addr: '%s' (%s)",
				c.addr, c.short)
			continue
		}

		if get, ok := storage.GetAddr(c.short); !ok || c.addr != get {
			t.Errorf("error get addr: short: %s, get %s, want %s",
				c.short, get, c.addr)
			continue
		}
	}
}
