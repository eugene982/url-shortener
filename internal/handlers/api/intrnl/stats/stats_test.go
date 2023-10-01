package stats

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type statsGetterFunc func() (int, int, error)

func (s statsGetterFunc) Stats(ctx context.Context) (int, int, error) {
	return s()
}

func TestStatsHandler(t *testing.T) {

	type want struct {
		code     int
		response string
	}
	type stats struct {
		err   error
		urls  int
		users int
	}

	tests := []struct {
		name  string
		stats stats
		want  want
	}{
		{
			name:  "request empty",
			stats: stats{nil, 0, 0},
			want:  want{200, `{"urls":0,"users":0}`},
		},
		{
			name:  "request err",
			stats: stats{errors.New("some error"), 0, 0},
			want:  want{500, "some error\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest("GET", "/internal/stats", nil)
			w := httptest.NewRecorder()

			stats := statsGetterFunc(func() (int, int, error) {
				return tt.stats.urls, tt.stats.users, tt.stats.err
			})

			NewStatsHandler(stats).ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			//
			assert.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.code == 200 {
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
			}

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.want.code == 200 {
				assert.JSONEq(t, tt.want.response, string(body))
			} else {
				assert.Equal(t, tt.want.response, string(body))
			}

		})
	}
}
