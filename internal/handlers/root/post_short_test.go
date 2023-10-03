package root

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eugene982/url-shortener/gen/go/proto"
	"github.com/eugene982/url-shortener/internal/middleware"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type setterFunc func() error

func (f setterFunc) Set(ctx context.Context, data model.StoreData) error {
	return f()
}

type shortenerFunc func(string) (string, error)

func (f shortenerFunc) Short(s string) (string, error) {
	return f(s)
}

func TestCreateShortHandler(t *testing.T) {

	type want struct {
		code  int
		short string
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "ok",
			body: "ya.ru",
			want: want{
				code:  201,
				short: "/YA.RU",
			},
		},
		{
			name: "conflict",
			body: "ya.ru",
			want: want{
				code:  409,
				short: "/YA.RU",
			},
		},
		{
			name: "err",
			body: "ya.ru",
			want: want{
				code:  404,
				short: "",
			},
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			r := httptest.NewRequest("GET", "/", strings.NewReader(tcase.body))
			w := httptest.NewRecorder()

			base := "/"

			setter := setterFunc(func() error {
				if tcase.want.code == 404 {
					return errors.New("some write error")
				} else if tcase.want.code == 409 {
					return storage.ErrAddressConflict
				}
				return nil
			})

			shorten := shortenerFunc(func(s string) (string, error) {
				return strings.ToUpper(s), nil
			})

			ru := middleware.RequestWithUserID(r, "user")
			NewCreateShortHandler(base, setter, shorten).ServeHTTP(w, ru)
			resp := w.Result()
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tcase.want.code, resp.StatusCode)

			if tcase.want.code != 404 {
				assert.Equal(t, tcase.want.short, string(body))
			}
		})
	}

}

func TestGRPCCreateShortHandler(t *testing.T) {

	type want struct {
		err   string
		short string
	}

	testErr := errors.New("some write error")

	tests := []struct {
		name string
		url  string
		err  error
		want want
	}{
		{
			name: "ok",
			url:  "ya.ru",
			want: want{
				err:   "",
				short: "/YA.RU",
			},
		},
		{
			name: "conflict",
			url:  "ya.ru",
			err:  storage.ErrAddressConflict,
			want: want{
				err:   storage.ErrAddressConflict.Error(),
				short: "",
			},
		},
		{
			name: "err",
			url:  "ya.ru",
			err:  testErr,
			want: want{
				err:   "",
				short: "",
			},
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			base := "/"

			setter := setterFunc(func() error {
				return tcase.err
			})

			shorten := shortenerFunc(func(s string) (string, error) {
				return strings.ToUpper(s), nil
			})

			in := proto.CreateShortRequest{
				User:        "user",
				OriginalUrl: tcase.url,
			}

			resp, err := NewGRPCCreateShortHandler(base, setter, shorten)(context.Background(), &in)
			if tcase.err == testErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.want.err, resp.Error)
			assert.Equal(t, tcase.want.short, resp.ShortUrl)
		})
	}

}
