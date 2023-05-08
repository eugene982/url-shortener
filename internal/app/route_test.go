// Тестирование обработки входящих запросов.

package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// простое хранилище
type mokStore struct {
	getAddrFunc  func(string) (string, bool)
	getShortFunc func(string) (string, bool)
	setFunc      func(string, string) bool
}

func (m mokStore) GetAddr(s string) (string, bool)  { return m.getAddrFunc(s) }
func (m mokStore) GetShort(s string) (string, bool) { return m.getShortFunc(s) }
func (m mokStore) Set(s1 string, s2 string) bool    { return m.setFunc(s1, s2) }

// простой сокращатель
type mokShorter func(string) string

func (m mokShorter) Short(s string) string { return m(s) }

func newTestApp(t *testing.T) *Application {
	st := mokStore{
		getShortFunc: func(string) (string, bool) { return "", false },
		getAddrFunc:  func(addr string) (string, bool) { return addr, addr != "" },
		setFunc:      func(s1, s2 string) bool { return s1 == s2 },
	}
	sh := mokShorter(func(addr string) string { return addr })

	a := NewApplication(sh, st, "")
	require.NotNil(t, a)
	return a
}

func TestRouterMethods(t *testing.T) {

	type want struct {
		code        int
		body        string
		contentType string
	}

	want404 := want{404, "404 page not found\n", "text/plain"}

	tests := []struct {
		method string
		want   want
	}{
		{method: http.MethodDelete, want: want404},
		{method: http.MethodConnect, want: want404},
		{method: http.MethodHead, want: want404},
		{method: http.MethodOptions, want: want404},
		{method: http.MethodPatch, want: want404},
		{method: http.MethodPut, want: want404},
		{method: http.MethodTrace, want: want404},
	}

	router := newTestApp(t).NewRouter()

	//ts := httptest.NewServer(app.NewRouter())

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {

			r := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Content-Type"), tt.want.contentType)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(body))
		})
	}
}

func TestRouterGet(t *testing.T) {

	type want struct {
		code     int
		Location string
	}

	tests := []struct {
		name string
		path string
		want want
	}{
		{"request uri /", "/", want{404, ""}},
		{"request ya.ru", "/ya.ru", want{307, "ya.ru"}},
		{"request yandex.ru", "/yandex.ru", want{307, "yandex.ru"}},
		// ...
	}

	router := newTestApp(t).NewRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			//
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Location"), tt.want.Location)
		})
	}
}

func TestRouterPost(t *testing.T) {

	type want struct {
		code int
		body string
	}
	type req struct {
		path        string
		contentType string
		short       string
	}

	tests := []struct {
		name string
		req  req
		want want
	}{
		{
			name: "request /",
			req:  req{"", "", ""},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request not text",
			req:  req{"", "application/json", ""},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request uri ya.ru",
			req:  req{"/ya.ru", "text/plain", ""},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request uri ya.ru",
			req:  req{"", "text/plain", "ya.ru"},
			want: want{201, "/ya.ru"},
		},
		{
			name: "request uri yandex.ru",
			req:  req{"", "text/plain;charset=utf-8", "yandex.ru"},
			want: want{201, "/yandex.ru"},
		},
		// ...
	}

	ts := httptest.NewServer(newTestApp(t).NewRouter())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, err := ts.Client().Post(ts.URL+tt.req.path,
				tt.req.contentType, strings.NewReader(tt.req.short))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.want.code == 201 {
				assert.Equal(t, ts.URL+tt.want.body, string(body))
			} else {
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}

func TestRouterPostJson(t *testing.T) {

	type want struct {
		code   int
		result string
	}
	type req struct {
		path        string
		contentType string
		short       string
	}

	tests := []struct {
		name string
		req  req
		want want
	}{
		{
			name: "request root",
			req:  req{"", "application/json", `{"url":"ya.ru}"}`},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request uri ya.ru",
			req:  req{"/api/shorten", "application/json", ""},
			want: want{404, "404 page not found\n"},
		},
		{
			name: "request uri ya.ru",
			req:  req{"/api/shorten", "application/json", `{"url":"ya.ru"}`},
			want: want{201, `{"result":"%s/ya.ru"}`},
		},
		{
			name: "request uri yandex.ru",
			req:  req{"/api/shorten", "application/json;charset=utf-8", `{"url":"yandex.ru"}`},
			want: want{201, `{"result":"%s/yandex.ru"}`},
		},
		// ...
	}

	ts := httptest.NewServer(newTestApp(t).NewRouter())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, err := ts.Client().Post(ts.URL+tt.req.path,
				tt.req.contentType, strings.NewReader(tt.req.short))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.want.code == 201 {
				assert.JSONEq(t, fmt.Sprintf(tt.want.result, ts.URL), string(body))
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
			} else {
				assert.Equal(t, tt.want.result, string(body))
			}
		})
	}
}
