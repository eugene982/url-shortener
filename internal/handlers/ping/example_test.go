package ping

import (
	"fmt"
	"io"
	"net/http/httptest"

	"github.com/eugene982/url-shortener/internal/handlers"
)

func ExampleNewPingHandler() {

	var pinger handlers.Pinger = PingerFunc(func() error {
		return nil
	})

	handler := NewPingHandler(pinger)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/ping", nil)
	handler.ServeHTTP(w, r)

	body, _ := io.ReadAll(w.Body)

	fmt.Println(w.Code)
	fmt.Println(string(body))

	// Output:
	// 200
	// pong

}
