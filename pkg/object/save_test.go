package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHAParser(t *testing.T) {
	m, err := getFiles("test", ".")
	assert.NoError(t, err)
	assert.NotNil(t, m)

	v, ok := m["save.go"]
	assert.True(t, ok)
	assert.Equal(t, "test-save.go", v)
}
