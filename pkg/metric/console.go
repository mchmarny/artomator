package metric

import (
	"context"
	"fmt"
)

func NewConsoleCounter() (Counter, error) {
	return &ConsoleCounter{}, nil
}

type ConsoleCounter struct {
}

func (r *ConsoleCounter) Count(ctx context.Context, metric string, count int64, labels map[string]string) error {
	fmt.Printf("console counter - %s:%d", metric, count)
	return nil
}
