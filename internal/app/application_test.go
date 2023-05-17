package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewApplication(t *testing.T) {
	a := NewApplication(nil, nil, "")
	assert.IsType(t, &Application{}, a)
}
