package shorten

import (
	"fmt"
	"io"
	"net/http/httptest"
	"strings"

	"github.com/eugene982/url-shortener/internal/middleware"
)

func ExampleNewShortenHandler() {

	base := "localhost:80/"

	setter := setterFunc(func() error {
		return nil
	})

	shortener := shortenerFunc(func(s string) (string, error) {
		return strings.ToUpper(s), nil
	})
	handler := NewShortenHandler(base, setter, shortener)

	reqbody := strings.NewReader(`{"url":"ya.ru"}`)
	r := httptest.NewRequest("POST", "/api/shorten", reqbody)
	r.Header.Set("Content-Type", "application/json")
	r = middleware.RequestWithUserID(r, "user")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	body, _ := io.ReadAll(w.Body)
	fmt.Println(w.Code)
	fmt.Println(string(body))

	// Output:
	// 201
	// {"result":"localhost:80/YA.RU"}

}
