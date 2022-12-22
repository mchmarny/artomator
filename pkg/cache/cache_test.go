package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryCache(t *testing.T) {
	ctx := context.Background()
	c := NewInMemoryCache()
	assert.NotNil(t, c)

	k := "key1"
	v := "val1"

	b, err := c.HasBeenProcessed(ctx, k, v)
	assert.NoError(t, err)
	assert.False(t, b)

	b, err = c.HasBeenProcessed(ctx, k, v)
	assert.NoError(t, err)
	assert.True(t, b)
}
