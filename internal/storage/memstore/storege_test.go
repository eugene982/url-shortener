// Тестирование хранилища

package memstore

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	require.NoError(t, err)

	_, err = New(testdata + "/err-url-db.json")
	require.Error(t, err)

	store, err := New(testdata + "/short-url-db.json")
	require.NoError(t, err)

	err = store.Close()
	require.NoError(t, err)
}

func TestPing(t *testing.T) {
	store, err := New("")
	require.NoError(t, err)

	ctx, close := context.WithCancel(context.Background())
	err = store.Ping(ctx)
	require.NoError(t, err)

	close()
	err = store.Ping(ctx)
	require.Error(t, err)
}

func TestGetAddr(t *testing.T) {
	store, err := New("")
	require.NoError(t, err)

	short := "short"
	want := "ya.ru"

	ctx, close := context.WithCancel(context.Background())
	err = store.Set(ctx, model.StoreData{
		OriginalURL: want,
		ShortURL:    short})
	require.NoError(t, err)

	get, err := store.GetAddr(ctx, short)
	require.NoError(t, err)
	require.Equal(t, want, get.OriginalURL)

	_, err = store.GetAddr(ctx, "-")
	require.ErrorIs(t, err, storage.ErrAddressNotFound)

	close()
	_, err = store.GetAddr(ctx, short)
	require.Error(t, err)
}

func TestSetAddr(t *testing.T) {
	store, err := New("")
	require.NoError(t, err)

	data := model.StoreData{
		OriginalURL: "ya.ru",
		ShortURL:    "short"}

	ctx, close := context.WithCancel(context.Background())
	err = store.Set(ctx, data)
	require.NoError(t, err)

	err = store.Set(ctx, data)
	require.ErrorIs(t, err, storage.ErrAddressConflict)

	err = store.Set(ctx, model.StoreData{})
	require.Error(t, err)

	close()
	err = store.Set(ctx, data)
	require.Error(t, err)

}

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
	require.NoError(t, err)

	ctx, close := context.WithCancel(context.Background())

	for _, c := range cases {
		data := []model.StoreData{
			{OriginalURL: c.addr, ShortURL: c.short},
		}

		err = storage.Update(ctx, data)
		if err != nil {
			t.Error(err)
		}

		get, err := storage.GetAddr(ctx, c.short)
		if err != nil {
			t.Error(err)
			continue
		}

		if c.addr != get.OriginalURL || c.short != get.ShortURL {
			t.Errorf("want: %s, %s; get %#v ", c.short, c.addr, get)
		}
	}

	close()
	err = storage.Update(ctx, nil)
	require.Error(t, err)
}

func TestGetUserURLs(t *testing.T) {
	store, err := New("")
	require.NoError(t, err)

	data := model.StoreData{
		UserID:      "user",
		ShortURL:    "short",
		OriginalURL: "ya.ru"}

	ctx, close := context.WithCancel(context.Background())
	err = store.Set(ctx, data)
	require.NoError(t, err)

	get, err := store.GetUserURLs(ctx, "user")
	require.NoError(t, err)
	require.Equal(t, 1, len(get))

	get, err = store.GetUserURLs(ctx, "-")
	require.NoError(t, err)
	require.Equal(t, 0, len(get))

	close()
	_, err = store.GetUserURLs(ctx, "-")
	require.Error(t, err)
}

func TestDeleteShort(t *testing.T) {
	store, err := New("")
	require.NoError(t, err)

	data := model.StoreData{
		UserID:      "user",
		ShortURL:    "short",
		OriginalURL: "ya.ru"}

	ctx, close := context.WithCancel(context.Background())
	err = store.Set(ctx, data)
	require.NoError(t, err)

	err = store.DeleteShort(ctx, []string{data.ShortURL})
	require.NoError(t, err)

	close()
	err = store.DeleteShort(ctx, nil)
	require.Error(t, err)
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

	err = storage.Set(ctx, model.StoreData{
		OriginalURL: addr,
		ShortURL:    addr,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = storage.GetAddr(ctx, addr)
		if err != nil {
			b.Fatal(err)
		}
	}

}

func BenchmarkUpdate(b *testing.B) {

	var (
		ctx  = context.Background()
		list = []model.StoreData{
			{
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
		err = storage.Update(ctx, list)
		if err != nil {
			b.Error(err)
		}
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

	err = storage.Update(ctx, []model.StoreData{
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
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// успокаиваем линтер
		_, err := storage.GetUserURLs(ctx, userID)
		if err != nil {
			b.Error(err)
		}
	}
}
