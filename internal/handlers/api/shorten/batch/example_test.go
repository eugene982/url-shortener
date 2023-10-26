package batch

import (
	"fmt"
	"io"
	"net/http/httptest"
	"strings"

	"github.com/eugene982/url-shortener/internal/middleware"
)

func ExampleNewBatchHandler() {

	base := "/"

	updater := updaterFunc(func() error {
		return nil
	})

	shorten := shortenerFunc(func(s string) (string, error) {
		return strings.ToUpper(s), nil
	})

	handler := NewBatchHandler(base, updater, shorten)

	reqbody := strings.NewReader(`[{"correlation_id":"1", "original_url":"ya.ru"}]`)
	r := httptest.NewRequest("POST", "/api/shorten/batch", reqbody)
	r.Header.Set("Content-Type", "application/json")
	r = middleware.RequestWithUserID(r, "user")

	w := httptest.NewRecorder()
	resp := w.Result()
	defer resp.Body.Close()

	handler.ServeHTTP(w, r)
	fmt.Println(w.Code)

	body, _ := io.ReadAll(w.Body)
	fmt.Println(string(body))

	// Output:
	// 201
	// [{"correlation_id":"1","short_url":"/YA.RU"}]
}
