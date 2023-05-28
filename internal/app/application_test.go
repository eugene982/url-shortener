package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplication(t *testing.T) {
	a, err := NewApplication(nil, nil, nil, "", "")
	require.NoError(t, err)
	assert.IsType(t, &Application{}, a)

	err = a.Close()
	require.NoError(t, err)
}
