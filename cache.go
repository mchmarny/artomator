package main

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	redis "github.com/go-redis/redis/v8"
)

const (
	expectedURIParts = 2
	// TODO: Externalize to allow deployment time configuration
	cacheExpireHrs = 168 // week
)

func getSHA(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) != expectedURIParts {
		return ""
	}
	return parts[1]
}

func keyBeenProcessed(ctx context.Context, k, v string) (bool, error) {
	if client == nil {
		return true, nil
	}
	_, err := client.Get(ctx, k).Result()
	if err == redis.Nil {
		err = client.Set(ctx, k, v, cacheExpireHrs*time.Hour).Err()
		if err != nil {
			return false, errors.Wrapf(err, "error setting key: %s - %v", k, err)
		}
		return false, nil
	} else if err != nil {
		return false, errors.Wrapf(err, "error getting key: %s - %v", k, err)
	}
	return true, nil
}
