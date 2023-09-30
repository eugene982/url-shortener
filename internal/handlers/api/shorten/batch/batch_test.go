package batch

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

type updaterFunc func() error

func (f updaterFunc) Update(ctx context.Context, list []model.StoreData) error {
	return f()
}

type shortenerFunc func(string) (string, error)

func (f shortenerFunc) Short(s string) (string, error) {
	return f(s)
}

func TestBatchHandler(t *testing.T) {

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
			req:  req{`[{"correlation_id":"", original_url:""}]`, "text/plain"},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request ya.ru",
			req:  req{`[{"correlation_id":"1", "original_url":"ya.ru"}]`, "application/json"},
			want: want{201, `[{"correlation_id":"1", "short_url":"/ya.ru"}]`},
		},
		{
			name: "request mail.ru and gmail.com",
			req: req{`[
					{"correlation_id":"2", "original_url":"mail.ru"},
					{"correlation_id":"3", "original_url":"gmail.com"}
				]`, "application/json"},

			want: want{201, `[
				{"correlation_id":"2", "short_url":"/mail.ru"},
				{"correlation_id":"3", "short_url":"/gmail.com"}
				]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest("POST", "/api/shorten/batch", strings.NewReader(tt.req.body))
			r.Header.Set("Content-Type", tt.req.contentType)
			w := httptest.NewRecorder()

			base := "/"

			shorten := func(s string) (string, error) {
				return s, nil
			}

			updater := func() error {
				return nil
			}

			ru := middleware.RequestWithUserID(r, "user")

			NewBatchHandler(base, updaterFunc(updater),
				shortenerFunc(shorten)).ServeHTTP(w, ru)
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
