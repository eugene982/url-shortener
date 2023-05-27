// Тестирование обработки входящих запросов.

package app

import (
	"bytes"
	"compress/gzip"
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
	getAddrFunc func(string) (string, bool)
	setFunc     func(string, string)
}

func (m mokStore) GetAddr(s string) (string, bool) { return m.getAddrFunc(s) }
func (m mokStore) Set(s1 string, s2 string)        { m.setFunc(s1, s2) }

// простой сокращатель
type mokShorter func(string) string

func (m mokShorter) Short(s string) string { return m(s) }

// простой логгер
type mokLogger struct{}

func (*mokLogger) Debug(msg string, a ...any) { fmt.Println(msg, a) }
func (*mokLogger) Info(msg string, a ...any)  { fmt.Println(msg, a) }
func (*mokLogger) Warn(msg string, a ...any)  { fmt.Println(msg, a) }
func (*mokLogger) Error(err error, a ...any)  { fmt.Println(err, a) }

// Тесты

func newTestApp(t *testing.T) *Application {
	st := mokStore{
		getAddrFunc: func(addr string) (string, bool) { return addr, addr != "" },
		setFunc:     func(s1, s2 string) {},
	}
	sh := mokShorter(func(addr string) string { return addr })
	logger := &mokLogger{}

	a := NewApplication(sh, st, logger, "")
	require.NotNil(t, a)

	return a
}

func newTestServer(t *testing.T) *httptest.Server {
	app := newTestApp(t)
	ts := httptest.NewServer(app.NewRouter())
	app.baseURL = ts.URL + "/"
	return ts
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

func TestRouterFindAddr(t *testing.T) {

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

func TestRouterCreateAddr(t *testing.T) {

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

	ts := newTestServer(t)

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

func TestRouterCreateApiShorten(t *testing.T) {

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

	router := newTestApp(t).NewRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(tt.req.body))
			r.Header.Set("Content-Type", tt.req.contentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)
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

func TestGzipCompression(t *testing.T) {
	app := newTestApp(t)
	handler := http.Handler(app.gzipMiddleware(http.HandlerFunc(app.createApiShorten)))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	// тело запроса
	requestBody := `{
        "url": "https://www.yandex.ru"
    }`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `{
    	"result": "/https://www.yandex.ru"
	}`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Content-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
