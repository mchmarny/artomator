package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	c := NewCommand("event", "echo")
	assert.NotNil(t, c)

	err := c.Run(context.TODO(), "hi")
	assert.NoError(t, err)
}

func TestBashCommand(t *testing.T) {
	c := NewBashCommand("event", "../../bin/test", "test")
	assert.NotNil(t, c)

	err := c.Run(context.TODO(), "hello")
	assert.NoError(t, err)
}
