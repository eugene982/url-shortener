package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplication(t *testing.T) {
	a := NewApplication(nil, nil)
	assert.IsType(t, &Application{}, a)
}

func TestAppRootHandler(t *testing.T) {

	type want struct {
		code        int
		response    string
		contentType string
	}

	want404 := want{404, "page not found", "text/plain"}

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

	a := NewApplication(nil, nil)
	require.NotNil(t, a)

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {

			assert.HTTPStatusCode(t, a.rootHandler, tt.method, "/", nil, tt.want.code)
			assert.HTTPBodyContains(t, a.rootHandler, tt.method, "/", nil, tt.want.response)

			r := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			a.rootHandler(w, r)
			res := w.Result()
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
		})
	}
}

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

func TestAppGetAddr(t *testing.T) {

	repiter := func(addr string) (string, bool) { return addr, addr != "" }
	a := NewApplication(nil, mokStore{getAddrFunc: repiter})

	require.NotNil(t, a)

	type want struct {
		ok       bool
		code     int
		Location string
	}

	tests := []struct {
		name string
		uri  string
		want want
	}{
		{"request uri /", "/", want{false, 200, ""}},
		{"request ya.ru", "/ya.ru", want{true, 307, "ya.ru"}},
		{"request yandex.ru", "/yandex.ru", want{true, 307, "yandex.ru"}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest("GET", tt.uri, nil)
			w := httptest.NewRecorder()

			get := a.getAddr(w, r)
			require.Equal(t, tt.want.ok, get)

			assert.Equal(t, tt.want.code, w.Code)
			assert.Contains(t, w.Header().Get("Location"), tt.want.Location)

		})
	}
}

func TestAppPostAddr(t *testing.T) {

	st := mokStore{
		getShortFunc: func(string) (string, bool) { return "", false },
		setFunc:      func(s1, s2 string) bool { return s1 == s2 },
	}
	sh := mokShorter(func(addr string) string { return addr })

	a := NewApplication(sh, st)
	require.NotNil(t, a)

	type want struct {
		ok   bool
		code int
		body string
	}
	type req struct {
		uri         string
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
			req:  req{"/", "", ""},
			want: want{false, 200, ""},
		},
		{
			name: "request not text",
			req:  req{"/", "application/json", ""},
			want: want{false, 200, ""},
		},
		{
			name: "request uri ya.ru",
			req:  req{"/ya.ru", "text/plain", ""},
			want: want{false, 200, ""},
		},
		{
			name: "request uri ya.ru",
			req:  req{"/", "text/plain", "ya.ru"},
			want: want{true, 201, "http://example.com/ya.ru"},
		},
		{
			name: "request uri yandex.ru",
			req:  req{"/", "text/plain;charset=utf-8", "yandex.ru"},
			want: want{true, 201, "http://example.com/yandex.ru"},
		},
		//

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest("POST", tt.req.uri, strings.NewReader(tt.req.short))
			r.Header.Set("Content-Type", tt.req.contentType)

			w := httptest.NewRecorder()
			get := a.postAddr(w, r)
			require.Equal(t, tt.want.ok, get)

			assert.Equal(t, tt.want.code, w.Code)
			body, err := w.Body.ReadString('\n')
			require.Equal(t, err, io.EOF)

			assert.Equal(t, body, tt.want.body)
		})
	}
}
