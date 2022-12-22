package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mchmarny/artomator/pkg/metric"
	"github.com/mchmarny/artomator/pkg/object"
	"github.com/mchmarny/artomator/pkg/pubsub"
	"github.com/pkg/errors"
)

const (
	CommandNameEvent = "event"
	ImageURISelf     = "us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator"
)

func (h *Handler) EventHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("processing event...")

	if err := h.Validate(CommandNameEvent); err != nil {
		log.Fatalf("service not configured")
	}

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

	if err := h.processEvent(r.Context(), e.Digest); err != nil {
		writeError(w, errors.Wrapf(err, "error processing event for %s", e.Digest))
		return
	}

	writeImageMessage(w, e.Digest, "event processed")
}

func (h *Handler) processEvent(ctx context.Context, digest string) error {
	log.Printf("processing digest: %s", digest)

	if strings.HasPrefix(digest, ImageURISelf) {
		log.Printf("digest of image used in this service, skilling: %s\n", digest)
		return nil
	}

	ri, err := getRegInfo(digest)
	if err != nil {
		return errors.Wrap(err, "error invoking caching service")
	}

	sha, err := parseSHA(digest)
	if err != nil {
		return errors.Wrap(err, "error parsing process event sha")
	}

	alreadyProcessed, err := h.cache.HasBeenProcessed(ctx, sha, digest)
	if err != nil {
		return errors.Wrap(err, "error invoking caching service")
	}

	if alreadyProcessed {
		log.Printf("image already processed: %s\n", digest)
		if err := h.counter.Count(ctx, metric.MakeMetricType("event/cached"), 1, ri); err != nil {
			log.Printf("unable to write metrics: %v", err)
		}
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

	if err := h.commands[CommandNameEvent].Run(ctx, digest, dir); err != nil {
		return errors.Wrap(err, "error executing command")
	}

	if h.bucket != "" {
		if err := object.Save(ctx, sha, h.bucket, dir); err != nil {
			return errors.Wrapf(err, "error saving %s resulting artifacts from: %s", sha, dir)
		}
	}

	if err := h.counter.Count(ctx, metric.MakeMetricType("event/processed"), 1, ri); err != nil {
		log.Printf("unable to write metrics: %v", err)
	}

	return nil
}
