package ping

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eugene982/url-shortener/internal/handlers"
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

func TestGRPCPingHandler(t *testing.T) {

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "ok", wantErr: false},
		{name: "internal error", wantErr: true},
	}

	ctx := context.Background()

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			var pinger handlers.Pinger = PingerFunc(func() error {
				if !tcase.wantErr {
					return nil
				} else {
					return fmt.Errorf("mock ping error")
				}
			})

			_, err := NewGRPCPingHandler(pinger)(ctx, nil)
			if tcase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
