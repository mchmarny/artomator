package main

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

const (
	testCommandName = "echo"
)

func execCmd(ctx context.Context, digest string) ([]byte, error) {
	if commandName == testCommandName {
		return []byte(commandName), nil
	}
	cmd := exec.CommandContext(ctx, "/bin/bash", commandName, digest, projectID, signKey)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrapf(err, "error executing command %s %s %s %s",
			commandName, digest, projectID, signKey)
	}
	log.Println(string(out))
	return out, nil
}
