package cmd

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Creates new instance of the command wrapper configured for bash.
func NewBashCommand(kind, name string, args ...string) *Command {
	return NewCommand(kind, "/bin/bash", append([]string{name}, args...)...)
}

// NewCommand creates new instance of the command wrapper.
func NewCommand(kind, name string, args ...string) *Command {
	c := &Command{
		Kind: kind,
		name: name,
		args: args,
	}
	return c
}

// Command is the command wrapper.
type Command struct {
	Kind string
	name string
	args []string
}

// Run executes the command.
func (c *Command) Run(ctx context.Context, args ...string) error {
	if c.args == nil {
		c.args = make([]string, 0)
	}

	a := append(c.args, args...)
	cc := exec.CommandContext(ctx, c.name, a...) //nolint
	cc.Stdout = os.Stdout
	cc.Stderr = os.Stdout

	if err := cc.Run(); err != nil {
		return errors.Wrapf(err, "error executing %s with %s", c.name, strings.Join(a, " "))
	}

	return nil
}
