package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"

	redis "github.com/go-redis/redis/v8"
)

const (
	cacheExpireHrs = 168 // week
)

func NewPersistedCache(ctx context.Context, ip, port string) (Cache, error) {
	log.Printf("connecting to redis %s:%s...\n", ip, port)
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, port),
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to redis using %s:%s", ip, port)
	}

	c := &PersistedCache{
		client: client,
	}
	return c, nil
}

type PersistedCache struct {
	client *redis.Client
}

func (c *PersistedCache) HasBeenProcessed(ctx context.Context, k, v string) (bool, error) {
	if c.client == nil {
		return true, nil
	}

	_, err := c.client.Get(ctx, k).Result()
	if errors.Is(err, redis.Nil) {
		err = c.client.Set(ctx, k, v, cacheExpireHrs*time.Hour).Err()
		if err != nil {
			return false, errors.Wrapf(err, "error setting key: %s - %v", k, err)
		}
		return false, nil
	} else if err != nil {
		return false, errors.Wrapf(err, "error getting key: %s - %v", k, err)
	}
	return true, nil
}
