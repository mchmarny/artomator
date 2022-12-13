package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/mchmarny/artomator/pkg/cache"
	"github.com/mchmarny/artomator/pkg/object"
	"github.com/mchmarny/artomator/pkg/pubsub"
)

const (
	expectedURIParts = 2
	actionInsert     = "INSERT"
	sigTagSuffix     = ".sig"
	attTagSuffix     = ".att"
)

type result struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewEventHandler(args []string, bucket string, cache cache.Cache) (*EventHandler, error) {
	h := &EventHandler{
		bucketName:   bucket,
		cacheService: cache,
		commandArgs:  args,
	}

	if err := h.validate(); err != nil {
		return nil, err
	}
	return h, nil
}

type EventHandler struct {
	bucketName   string
	commandArgs  []string
	cacheService cache.Cache
}

func (h *EventHandler) validate() error {
	if h.commandArgs == nil {
		return errors.New("invalid command args")
	}
	if h.cacheService == nil {
		return errors.New("cache service is required")
	}
	return nil
}

func (h *EventHandler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("processing event...")

	if err := h.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	e, err := pubsub.ParseEvent(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	sha, err := parseSHA(e.Digest)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	log.Printf("event action: %s, digest: %s, tag: %s, sha:%s\n",
		e.Action, e.Digest, e.Tag, sha)

	if e.Action != actionInsert {
		writeMessage(w, fmt.Sprintf("unsupported event type: %s", e.Action))
		return
	}

	if strings.HasSuffix(e.Tag, sigTagSuffix) ||
		strings.HasSuffix(e.Tag, attTagSuffix) {
		writeMessage(w, fmt.Sprintf("signature or attestation event: %s", e.Tag))
		return
	}

	alreadyProcessed, err := h.cacheService.HasBeenProcessed(r.Context(), sha, e.Digest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if alreadyProcessed {
		writeMessage(w, fmt.Sprintf("image already processed: %s", e.Digest))
		return
	}

	c, err := makeFolder(sha)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			errors.Wrapf(err, "error creating context from sha: %s", sha))
		return
	}
	defer func() {
		if err = os.RemoveAll(c); err != nil {
			log.Printf("error deleting context: %s", c)
		}
	}()

	out, err := runCommand(r.Context(), append(h.commandArgs, e.Digest, c))
	if err != nil {
		writeError(w, http.StatusInternalServerError, errors.Wrap(err, "error executing"))
		return
	}

	if h.bucketName != "" {
		if err := object.Save(h.bucketName, c); err != nil {
			log.Printf("error saving resulting artifacts from: %s", c)
		}
	}

	log.Printf("done: %s\n", string(out))
	writeMessage(w, "ok")
}

func writeError(w http.ResponseWriter, s int, err error) {
	log.Println(err)
	w.WriteHeader(s)
	if err = json.NewEncoder(w).Encode(result{
		Status: s,
		Error:  err.Error(),
	}); err != nil {
		log.Printf("error encoding: %s (%d) - %v", err.Error(), s, err)
	}
}

func writeMessage(w http.ResponseWriter, msg string) {
	log.Println(msg)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result{
		Status:  http.StatusOK,
		Message: msg,
	}); err != nil {
		log.Printf("error encoding: %s - %v", msg, err)
	}
}

func parseSHA(uri string) (string, error) {
	parts := strings.Split(uri, ":")
	if len(parts) != expectedURIParts {
		return "", errors.Errorf("unable to parse SHA from %s", uri)
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
