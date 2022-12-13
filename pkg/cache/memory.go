package cache

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var (
	lock sync.Mutex
)

func NewInMemoryCache() Cache {
	c := &InMemoryCache{
		data: make(map[string]string),
	}
	return c
}

type InMemoryCache struct {
	data map[string]string
}

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
