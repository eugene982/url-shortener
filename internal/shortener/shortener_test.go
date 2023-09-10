package shortener

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShort(t *testing.T) {
	shortener := NewSimpleShortener()

	tests := []struct {
		name      string
		addr      string
		wantShort string
	}{
		{
			name:      "error",
			addr:      "",
			wantShort: "0000000000",
		},
		{
			name:      "error",
			addr:      "http://ya.ru",
			wantShort: "SbXCfyuJdo",
		},
		{
			name:      "error",
			addr:      "http://go.gl",
			wantShort: "q5fgk4YdbB",
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			short, err := shortener.Short(tcase.addr)
			require.NoError(t, err)
			assert.Equal(t, tcase.wantShort, short)

		})
	}
}

func BenchmarkShort(b *testing.B) {
	shortener := NewSimpleShortener()
	b.ResetTimer()

	b.Run("ya", func(b *testing.B) {
		addr := "https://yandex.ru/search/?text=go+benchmark"
		for i := 0; i < b.N; i++ {
			_, err := shortener.Short(addr)
			if err != nil {
				b.Error(err)
			}
		}
	})

	b.Run("google", func(b *testing.B) {
		addr := "https://www.google.com/search?q=go+benchmark&oq=go+benchmark"
		for i := 0; i < b.N; i++ {
			_, err := shortener.Short(addr)
			if err != nil {
				b.Error(err)
			}
		}
	})
}
