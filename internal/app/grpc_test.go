package app

import (
	"context"
	"testing"

	"github.com/eugene982/url-shortener/gen/go/proto/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCServer(t *testing.T) {

	testapp := newTestApp(t)
	server, err := NewGRPCServer(testapp, ":8083")
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("ping", func(t *testing.T) {
		resp, err := server.proto.Ping(ctx, nil)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("find", func(t *testing.T) {
		resp, err := server.proto.FindAddr(ctx, &proto.FindAddrRequest{ShortUrl: "ya.ru"})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("short", func(t *testing.T) {
		resp, err := server.proto.CreateShort(ctx, &proto.CreateShortRequest{
			OriginalUrl: "ya.ru"})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("batch", func(t *testing.T) {
		resp, err := server.proto.BatchShort(ctx, &proto.BatchRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("urls", func(t *testing.T) {
		resp, err := server.proto.GetUserURLs(ctx, &proto.UserURLsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("del urls", func(t *testing.T) {
		resp, err := server.proto.DelUserURLs(ctx, &proto.DelUserURLsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

}
