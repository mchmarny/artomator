package cache

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var (
	lock sync.Mutex
)

// NewInMemoryCache creates new in-memory cache.
func NewInMemoryCache() Cache {
	c := &InMemoryCache{
		data: make(map[string]string),
	}
	return c
}

// InMemoryCache is the in-memory cache implementation.
type InMemoryCache struct {
	data map[string]string
}

// HasBeenProcessed checks if the key has been processed before.
func (c *InMemoryCache) HasBeenProcessed(ctx context.Context, k, v string) (bool, error) {
	if k == "" || v == "" {
		return false, errors.New("neither, key (k) and value (v) can be empty.")
	}

	if _, ok := c.data[k]; ok {
		return true, nil
	}

	lock.Lock()
	defer lock.Unlock()
	c.data[k] = v
	return false, nil
}
