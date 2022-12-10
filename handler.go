package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func writeMessage(w http.ResponseWriter, s int, m string) {
	w.WriteHeader(s)
	if err := json.NewEncoder(w).Encode(result{Status: s, Message: m}); err != nil {
		log.Printf("error encoding message: %s (%d) - %v", m, s, err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("processing event...")

	e, err := getEvent(r)
	if err != nil {
		writeMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	imgSHA := getSHA(e.Digest)
	fmt.Printf("event action: %s, digest: %s, tag: %s, sha:%s\n", e.Action, e.Digest, e.Tag, imgSHA)

	if e.Action != actionInsert {
		writeMessage(w, http.StatusOK, fmt.Sprintf("unsupported event type: %s", e.Action))
		return
	}

	if strings.HasSuffix(e.Tag, sigTagSuffix) || strings.HasSuffix(e.Tag, attTagSuffix) {
		writeMessage(w, http.StatusOK, fmt.Sprintf("signature or attestation event: %s", e.Tag))
		return
	}

	k, err := getCachedKey(r.Context(), imgSHA, e.Digest)
	if err != nil {
		writeMessage(w, http.StatusInternalServerError,
			fmt.Sprintf("error getting cached key: %v", err))
		return
	}

	if k != e.Digest {
		writeMessage(w, http.StatusInternalServerError,
			fmt.Sprintf("same SHA (%s) diff URI (cached:%s, request:%s)", imgSHA, k, e.Digest))
		return
	}

	out, err := execCmd(r.Context(), e.Digest)
	if err != nil {
		writeMessage(w, http.StatusInternalServerError,
			fmt.Sprintf("error executing command: %v", err))
		return
	}

	log.Printf("done: %s\n", string(out))
	writeMessage(w, http.StatusOK, "ok")
}
