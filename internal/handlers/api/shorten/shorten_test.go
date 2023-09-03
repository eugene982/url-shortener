package shorten

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eugene982/url-shortener/internal/middleware"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type baseURLGetterFunc func() string

func (f baseURLGetterFunc) GetBaseURL() string {
	return f()
}

type setterFunc func() error

func (f setterFunc) Set(ctx context.Context, data model.StoreData) error {
	return f()
}

type shortenerFunc func(string) (string, error)

func (f shortenerFunc) Short(s string) (string, error) {
	return f(s)
}

func TestRouterHandlerApiShorten(t *testing.T) {

	type want struct {
		code     int
		response string
	}
	type req struct {
		body        string
		contentType string
	}

	tests := []struct {
		name string
		req  req
		want want
	}{
		{
			name: "request empty",
			req:  req{"", ""},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request empy json",
			req:  req{`{"url":""}`, "application/json"},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request not json",
			req:  req{`{"url":"ya.ru"}`, "text/plain"},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request uri ya.ru",
			req:  req{`{"url":"ya.ru"}`, "application/json"},
			want: want{201, `{"result":"/ya.ru"}`},
		},
		{
			name: "request uri yandex.ru",
			req:  req{`{"url":"yandex.ru"}`, "application/json;charset=utf-8"},
			want: want{201, `{"result":"/yandex.ru"}`},
		},
		// ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(tt.req.body))
			r.Header.Set("Content-Type", tt.req.contentType)
			w := httptest.NewRecorder()

			base := baseURLGetterFunc(func() string {
				return "/"
			})

			setter := setterFunc(func() error {
				return nil
			})

			shortener := shortenerFunc(func(s string) (string, error) {
				return s, nil
			})

			ru := middleware.RequestWithUserID(r, "user")
			NewShortenHandler(base, setter, shortener).ServeHTTP(w, ru)
			resp := w.Result()
			defer resp.Body.Close()
			//
			assert.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.code == 201 {
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
			}

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.want.code == 201 {
				assert.JSONEq(t, tt.want.response, string(body))
			} else {
				assert.Equal(t, tt.want.response, string(body))
			}

		})
	}
}
