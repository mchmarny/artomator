package cache

import (
	"context"
)

type Cache interface {
	HasBeenProcessed(ctx context.Context, k, v string) (bool, error)
}
