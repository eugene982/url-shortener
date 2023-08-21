package root

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/eugene982/url-shortener/internal/model"
	"github.com/stretchr/testify/assert"
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
		want want
	}{
		{"request uri /", "/", want{410, ""}},
		{"request ya.ru", "/ya.ru", want{307, "ya.ru"}},
		{"request yandex.ru", "/yandex.ru", want{307, "yandex.ru"}},
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
					DeletedFlag: tt.want.code == 410}, nil
			})

			NewFindAddrHandler(getter).ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Location"), tt.want.Location)
		})
	}
}
