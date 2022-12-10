package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type result struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func writeError(w http.ResponseWriter, s int, err error) {
	log.Println(err)
	w.WriteHeader(s)
	if err := json.NewEncoder(w).Encode(result{Status: s, Error: err.Error()}); err != nil {
		log.Printf("error encoding message: %s (%d) - %v", err.Error(), s, err)
	}
}

func writeMessage(w http.ResponseWriter, s int, m string) {
	log.Println(m)
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
		writeError(w, http.StatusBadRequest, err)
		return
	}

	imgSHA := getSHA(e.Digest)
	log.Printf("event action: %s, digest: %s, tag: %s, sha:%s\n", e.Action, e.Digest, e.Tag, imgSHA)

	if e.Action != actionInsert {
		writeMessage(w, http.StatusOK, fmt.Sprintf("unsupported event type: %s", e.Action))
		return
	}

	if strings.HasSuffix(e.Tag, sigTagSuffix) || strings.HasSuffix(e.Tag, attTagSuffix) {
		writeMessage(w, http.StatusOK, fmt.Sprintf("signature or attestation event: %s", e.Tag))
		return
	}

	alreadyProcessed, err := keyBeenProcessed(r.Context(), imgSHA, e.Digest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if alreadyProcessed {
		writeMessage(w, http.StatusOK, fmt.Sprintf("image already processed: %s", e.Digest))
		return
	}

	out, err := execCmd(r.Context(), e.Digest)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Errorf("error executing command: %v", err))
		return
	}

	log.Printf("done: %s\n", string(out))
	writeMessage(w, http.StatusOK, "ok")
}
