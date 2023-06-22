// Тестирование хранилища

package memstore

import (
	"context"
	"testing"

	"github.com/eugene982/url-shortener/internal/model"
)

func TestUpdateAddr(t *testing.T) {

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

	storage, err := New("")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	for _, c := range cases {
		data := []model.StoreData{
			{OriginalURL: c.addr, ShortURL: c.short},
		}

		storage.Update(ctx, data)

		get, err := storage.GetAddr(ctx, c.short)
		if err != nil {
			t.Error(err)
			continue
		}

		if c.addr != get.OriginalURL || c.short != get.ShortURL {
			t.Errorf("want: %s, %s; get %#v ", c.short, c.addr, get)
		}
	}
}
