package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func execCmd(ctx context.Context, digest string) ([]byte, error) {
	if commandName == "test" {
		return []byte(commandName), nil
	}
	cmd := exec.CommandContext(ctx, "/bin/bash", commandName, digest, projectID, signKey)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "error executing command %s %s %s %s",
			commandName, digest, projectID, signKey)
	}
	return out, nil
}
