package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidRequestShorten(t *testing.T) {

	testCases := []struct {
		name    string
		request RequestShorten
		wantRes bool
		wantErr bool
	}{
		{
			name:    "empty",
			request: RequestShorten{},
			wantRes: false,
			wantErr: true,
		},
		{
			name:    "has url",
			request: RequestShorten{URL: "ya.ru"},
			wantRes: true,
			wantErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ok, err := tC.request.IsValid()
			assert.Equal(t, tC.wantRes, ok)

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidStoreData(t *testing.T) {

	testCases := []struct {
		name    string
		request StoreData
		wantRes bool
		wantErr bool
	}{
		{
			name:    "empty",
			request: StoreData{},
			wantRes: false,
			wantErr: true,
		},
		{
			name:    "empty short",
			request: StoreData{OriginalURL: "ya.ru"},
			wantRes: false,
			wantErr: true,
		},
		{
			name:    "empty original",
			request: StoreData{ShortURL: "short"},
			wantRes: false,
			wantErr: true,
		},
		{
			name:    "ok",
			request: StoreData{ShortURL: "short", OriginalURL: "ya.ry"},
			wantRes: true,
			wantErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ok, err := tC.request.IsValid()
			assert.Equal(t, tC.wantRes, ok)

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidBatchRequest(t *testing.T) {

	testCases := []struct {
		name    string
		request BatchRequest
		wantRes bool
		wantErr bool
	}{
		{
			name:    "empty",
			request: BatchRequest{},
			wantRes: false,
			wantErr: true,
		},
		{
			name:    "empty correlation",
			request: BatchRequest{OriginalURL: "ya.ru"},
			wantRes: false,
			wantErr: true,
		},
		{
			name:    "empty original",
			request: BatchRequest{CorrelationID: "short"},
			wantRes: false,
			wantErr: true,
		},
		{
			name:    "ok",
			request: BatchRequest{CorrelationID: "short", OriginalURL: "ya.ry"},
			wantRes: true,
			wantErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ok, err := tC.request.IsValid()
			assert.Equal(t, tC.wantRes, ok)

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
