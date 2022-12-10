package main

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	redis "github.com/go-redis/redis/v8"
)

const expectedURIParts = 2

func getSHA(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) != expectedURIParts {
		return ""
	}
	return parts[1]
}

func getCachedKey(ctx context.Context, key, val string) (string, error) {
	if client == nil {
		return val, nil
	}
	v, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		err = client.Set(ctx, key, val, 0).Err()
		if err != nil {
			return "", errors.Wrapf(err, "error setting key: %s - %v", key, err)
		}
		v = val
	} else if err != nil {
		return "", errors.Wrapf(err, "error getting key: %s - %v", key, err)
	}
	return v, nil
}
