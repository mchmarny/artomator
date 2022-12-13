package cache

import (
	"context"
	"testing"
)

func TestInMemoryCache(t *testing.T) {
	ctx := context.Background()
	c := NewInMemoryCache()
	k := "key1"
	v := "val1"
	b, err := c.HasBeenProcessed(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Fatal("new key shouldn't exist")
	}
	b, err = c.HasBeenProcessed(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("new key should have existed")
	}
}
