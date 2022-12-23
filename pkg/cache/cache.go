package cache

import (
	"context"
)

// Cache is the interface for cache implementations.
type Cache interface {
	HasBeenProcessed(ctx context.Context, k, v string) (bool, error)
}
