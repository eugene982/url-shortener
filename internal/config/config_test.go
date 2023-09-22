package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonConfig(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	require.NoError(t, err)

	err = decodeJsonConfigFile(testdata + "/nullfile.json")
	require.Error(t, err)

	err = decodeJsonConfigFile(testdata + "/errconf.json")
	require.Error(t, err)

	err = decodeJsonConfigFile(testdata + "/config.json")
	require.NoError(t, err)

	assert.Equal(t, "localhost:8080", config.ServAddr)
	assert.Equal(t, "http://localhost", config.BaseURL)
	assert.Equal(t, "/path/to/file.db", config.FileStoragePath)
	assert.Equal(t, "postgres://", config.DatabaseDSN)
	assert.Equal(t, true, config.EnableHTTPS)
}

func TestConfig(t *testing.T) {
	_, err := Config()
	require.NoError(t, err)
}
