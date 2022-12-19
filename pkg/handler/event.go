package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mchmarny/artomator/pkg/object"
	"github.com/mchmarny/artomator/pkg/pubsub"
	"github.com/pkg/errors"
)

func (h *EventHandler) EventHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("processing event...")

	if r.Method != http.MethodPost {
		writeError(w, errors.Errorf("method %s not supported, expected POST", r.Method))
		return
	}

	e, err := pubsub.ParseEvent(r)
	if err != nil {
		writeError(w, errors.New("error parsing event"))
		return
	}

	log.Printf("event action: %s, digest: %s, tag: %s", e.Action, e.Digest, e.Tag)

	if e.Action != actionInsert {
		log.Printf("unsupported event type: %s\n", e.Action)
		writeMessage(w, successResponseMessage)
		return
	}

	if strings.HasSuffix(e.Tag, sigTagSuffix) ||
		strings.HasSuffix(e.Tag, attTagSuffix) {
		log.Printf("signature or attestation event: %s\n", e.Tag)
		writeMessage(w, successResponseMessage)
		return
	}

	if err := h.processEvent(r.Context(), e.Digest, h.eventCmdArgs); err != nil {
		writeError(w, errors.Wrapf(err, "error processing event for %s", e.Digest))
		return
	}

	writeImageMessage(w, e.Digest, "event processed")
}

func (h *EventHandler) processEvent(ctx context.Context, digest string, args []string) error {
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
	if err := runCommand(ctx, cmdArgs); err != nil {
		return errors.Wrapf(err, "error executing command: %s\n", strings.Join(cmdArgs, ","))
	}

	if h.bucketName != "" {
		if err := object.Save(ctx, sha, h.bucketName, dir); err != nil {
			return errors.Wrapf(err, "error saving %s resulting artifacts from: %s", sha, dir)
		}
	}

	return nil
}
