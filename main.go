package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const (
	actionInsert = "INSERT"
	commandName  = "automator"
	portDefault  = "8080"
)

func main() {
	http.HandleFunc("/", scriptHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = portDefault
		fmt.Printf("using default port %s\n", port)
	}
	address := fmt.Sprintf(":%s", port)

	fmt.Printf("starting server %s\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		panic(err)
	}
}

type event struct {
	Action string `json:"action"`
	Digest string `json:"digest"`
	Tag    string `json:"tag"`
}

type message struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func writeMessage(w http.ResponseWriter, s int, m string) {
	w.WriteHeader(s)
	if err := json.NewEncoder(w).Encode(message{Status: s, Message: m}); err != nil {
		log.Printf("error encoding message: %s (%d) - %v", m, s, err)
	}
}

func scriptHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var e event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeMessage(w, http.StatusInternalServerError, fmt.Sprintf("error parsing event: %v", err))
		return
	}

	fmt.Printf("action: %s, digest: %s, tag: %s\n", e.Action, e.Digest, e.Tag)

	if e.Action != actionInsert {
		writeMessage(w, http.StatusOK, fmt.Sprintf("unsupported event type: %s", e.Action))
		return
	}

	if e.Tag != "" {
		writeMessage(w, http.StatusOK, fmt.Sprintf("unsupported insert type: %s - %s", e.Action, e.Tag))
		return
	}

	cmd := exec.CommandContext(r.Context(), "/bin/bash", commandName, e.Digest) //nolint:gosec
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		writeMessage(w, http.StatusInternalServerError, fmt.Sprintf("error executing command: %v", err))
		return
	}

	fmt.Printf("command output: %s\n", out)

	writeMessage(w, http.StatusOK, "ok")
}
