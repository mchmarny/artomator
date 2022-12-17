package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/mchmarny/artomator/pkg/object"
)

const (
	expectedURIParts    = 2
	actionInsert        = "INSERT"
	sigTagSuffix        = ".sig"
	attTagSuffix        = ".att"
	sbomFormatParamName = "format"
)

func (h *EventHandler) ProcessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("processing event...")

	if r.Method != http.MethodPost {
		writeError(w, errors.Errorf("method %s not supported, expected POST", r.Method))
		return
	}

	digest := r.URL.Query().Get(imageDigestQueryParamName)
	if digest == "" {
		writeError(w, errors.Errorf("process %s parameter not set", imageDigestQueryParamName))
		return
	}

	if err := h.process(r.Context(), digest, h.processCmdArgs); err != nil {
		writeError(w, err)
		return
	}
	writeMessage(w, "processed")
}

func (h *EventHandler) process(ctx context.Context, digest string, args []string) error {
	log.Printf("processing digest: %s", digest)

	sha, err := parseSHA(digest)
	if err != nil {
		return errors.Wrap(err, "error parsing process event sha")
	}

	alreadyProcessed, err := h.cacheService.HasBeenProcessed(ctx, sha, digest)
	if err != nil {
		return errors.Wrap(err, "error invoking caching service")
	}

	if alreadyProcessed {
		log.Printf("image already processed: %s\n", digest)
		return nil
	}

	dir, err := makeFolder(sha)
	if err != nil {
		return errors.Wrapf(err, "error creating context from sha: %s", sha)
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			log.Printf("error deleting context: %s\n", dir)
		}
	}()

	cmdArgs := append(args, digest, dir)
	out, err := runCommand(ctx, cmdArgs)
	if err != nil {
		return errors.Wrapf(err, "error executing command: %s\n", strings.Join(cmdArgs, ","))
	}

	if h.bucketName != "" {
		if err := object.Save(ctx, sha, h.bucketName, dir); err != nil {
			return errors.Wrapf(err, "error saving %s resulting artifacts from: %s", sha, dir)
		}
	}

	log.Printf("done: %s\n", string(out))
	return nil
}

// parseSHA ensures that the image URI is actually a digest
// shouldn't process based on labels
// example: "us-west1-docker.pkg.dev/test/test/tester@sha256:123"
func parseSHA(uri string) (string, error) {
	parts := strings.Split(uri, "@")
	if len(parts) != expectedURIParts {
		return "", errors.Errorf("unable to parse digest (@) from %s", uri)
	}

	parts = strings.Split(parts[1], ":")
	if len(parts) != expectedURIParts {
		return "", errors.Errorf("unable to parse SHA (:) from %s", uri)
	}
	return parts[1], nil
}

func runCommand(ctx context.Context, args []string) ([]byte, error) {
	c := exec.CommandContext(ctx, "/bin/bash", args...)
	c.Stderr = os.Stderr
	out, err := c.Output()
	if err != nil {
		return nil, errors.Wrapf(err, "error executing: %s",
			strings.Join(args, " "))
	}
	log.Println(string(out))
	return out, nil
}

func makeFolder(sha string) (string, error) {
	p := fmt.Sprintf("./%s", sha)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(p, os.ModePerm); err != nil {
			return "", errors.Wrap(err, "error creating folder")
		}
	}
	return p, nil
}
