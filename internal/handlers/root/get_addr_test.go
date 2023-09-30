package root

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/eugene982/url-shortener/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type addGetterFunc func() (model.StoreData, error)

func (g addGetterFunc) GetAddr(context.Context, string) (model.StoreData, error) {
	return g()
}

func TestFindAddrHandler(t *testing.T) {

	type want struct {
		code     int
		Location string
	}

	tests := []struct {
		name string
		path string
		err  error
		want want
	}{
		{"request uri /", "/", nil, want{410, ""}},
		{"request ya.ru", "/ya.ru", nil, want{307, "ya.ru"}},
		{"request yandex.ru", "/yandex.ru", nil, want{307, "yandex.ru"}},
		{
			"request 404",
			"/",
			errors.New("some err"),
			want{404, ""},
		},
		{
			"request not found",
			"/",
			storage.ErrAddressNotFound,
			want{404, ""},
		},
		// ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			getter := addGetterFunc(func() (model.StoreData, error) {
				return model.StoreData{
					ShortURL:    tt.path,
					OriginalURL: tt.want.Location,
					DeletedFlag: tt.want.code == 410}, tt.err
			})

			NewFindAddrHandler(getter).ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Location"), tt.want.Location)
		})
	}
}

func TestGRPCFindAddrHandler(t *testing.T) {

	testErr := errors.New("some err")

	type want struct {
		err      string
		location string
	}

	tests := []struct {
		name string
		path string
		err  error
		want want
	}{
		{"request uri /", "/", nil, want{"Delete", ""}},
		{"request ya.ru", "/ya.ru", nil, want{"", "ya.ru"}},
		{"request yandex.ru", "/yandex.ru", nil, want{"", "yandex.ru"}},
		{
			"request 404",
			"/",
			testErr,
			want{"", ""},
		},
		{
			"request not found /",
			"/",
			storage.ErrAddressNotFound,
			want{storage.ErrAddressNotFound.Error(), ""},
		},
		// ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getter := addGetterFunc(func() (model.StoreData, error) {
				return model.StoreData{
					ShortURL:    tt.path,
					OriginalURL: tt.want.location,
					DeletedFlag: tt.want.err == "Delete"}, tt.err
			})

			in := proto.FindAddrRequest{
				ShortUrl: tt.path,
			}

			resp, err := NewGRPCFindAddrHandler(getter)(context.Background(), &in)
			if tt.err == testErr {
				assert.Equal(t, err, tt.err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.err, resp.Error)
			assert.Contains(t, tt.want.location, resp.OriginalUrl)
		})
	}
}
