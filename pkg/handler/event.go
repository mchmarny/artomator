package handler

import (
	"log"
	"net/http"
	"strings"

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

	if err := h.process(r.Context(), e.Digest, h.processCmdArgs); err != nil {
		writeError(w, errors.Wrapf(err, "error processing event for %s", e.Digest))
		return
	}

	writeImageMessage(w, e.Digest, "event processed")
}
