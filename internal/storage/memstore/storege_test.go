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

func BenchmarkGetAddr(b *testing.B) {

	var (
		ctx  = context.Background()
		addr = "http://yandex.ru"
	)

	storage, err := New("")
	if err != nil {
		b.Fatal(err)
	}

	storage.Set(ctx, model.StoreData{
		OriginalURL: addr,
		ShortURL:    addr,
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.GetAddr(ctx, addr)
	}

}

func BenchmarkUpdate(b *testing.B) {

	var (
		ctx  = context.Background()
		list = []model.StoreData{
			model.StoreData{
				OriginalURL: "http://yandex.ru",
				ShortURL:    "http://yandex.ru",
			},
		}
	)

	storage, err := New("")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.Update(ctx, list)
	}
}

func BenchmarkGetUserURLs(b *testing.B) {

	var (
		ctx    = context.Background()
		userID = "user"
	)

	storage, err := New("")
	if err != nil {
		b.Fatal(err)
	}

	storage.Update(ctx, []model.StoreData{
		{
			UserID:      userID,
			OriginalURL: "http://yandex.ru",
			ShortURL:    "http://yandex.ru",
		},
		{
			UserID:      userID,
			OriginalURL: "http://google.com",
			ShortURL:    "http://google.com",
		},
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.GetUserURLs(ctx, userID)
	}

}
