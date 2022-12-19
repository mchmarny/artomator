package cmd

import (
	"context"
	"testing"
)

func TestCommand(t *testing.T) {
	c := NewCommand("event", "echo")
	if c == nil {
		t.Fatal()
	}
	if err := c.Run(context.TODO(), "hi"); err != nil {
		t.Fatal(err)
	}
}

func TestBashCommand(t *testing.T) {
	c := NewBashCommand("event", "../../bin/test", "test")
	if c == nil {
		t.Fatal()
	}
	if err := c.Run(context.TODO(), "hello"); err != nil {
		t.Fatal(err)
	}
}
