package ping

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/eugene982/url-shortener/internal/handlers"
	"github.com/stretchr/testify/assert"
)

type PingerFunc func() error

func (fn PingerFunc) Ping(context.Context) error {
	return fn()
}

func TestPingHandler(t *testing.T) {

	tests := []struct {
		name       string
		wantStatus int
	}{
		{name: "ok", wantStatus: 200},
		{name: "internal error", wantStatus: 500},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/ping", nil)

			var pinger handlers.Pinger = PingerFunc(func() error {
				if tcase.wantStatus == 200 {
					return nil
				} else {
					return fmt.Errorf("mock ping error")
				}
			})

			NewPingHandler(pinger).ServeHTTP(w, r)
			assert.Equal(t, tcase.wantStatus, w.Code)
		})
	}
}
