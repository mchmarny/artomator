package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const (
	actionInsert     = "INSERT"
	commandName      = "artomator"
	portDefault      = "8080"
	testSubscription = "test"
)

type pubsubMessage struct {
	Message      message `json:"message"`
	Subscription string  `json:"subscription"`
}

type message struct {
	Data []byte `json:"data,omitempty"`
	ID   string `json:"id"`
}

type event struct {
	Action string `json:"action"`
	Digest string `json:"digest"`
	Tag    string `json:"tag"`
}

type result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func writeMessage(w http.ResponseWriter, s int, m string) {
	w.WriteHeader(s)
	if err := json.NewEncoder(w).Encode(result{Status: s, Message: m}); err != nil {
		log.Printf("error encoding message: %s (%d) - %v", m, s, err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var m pubsubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error parsing pubsub message: %v", err))
		return
	}

	d, err := base64.StdEncoding.DecodeString(string(m.Message.Data))
	if err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error decoding message data: %v", err))
		return
	}
	fmt.Printf("message data: %s\n", d)

	var e event
	if err = json.Unmarshal(d, &e); err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error parsing event: %v", err))
		return
	}
	fmt.Printf("event action: %s, digest: %s, tag: %s\n", e.Action, e.Digest, e.Tag)

	if e.Action != actionInsert || e.Tag != "" {
		writeMessage(w, http.StatusNoContent, fmt.Sprintf("unsupported event type: %s", e.Action))
		return
	}

	if m.Subscription == testSubscription {
		fmt.Println("skipping executing command during test")
		writeMessage(w, http.StatusAccepted, "ok")
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

func main() {
	http.HandleFunc("/", handler)
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
