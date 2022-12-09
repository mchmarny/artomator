package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"google.golang.org/api/pubsub/v1"
)

const (
	actionInsert     = "INSERT"
	commandName      = "artomator"
	portDefault      = "8080"
	testSubscription = "test"
)

var (
	projectID = os.Getenv("PROJECT_ID")
	signKey   = os.Getenv("SIGN_KEY")
)

type pubsubMessage struct {
	Message      pubsub.PubsubMessage `json:"message"`
	Subscription string               `json:"subscription"`
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
	log.Println("processing event...")

	var m pubsubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error parsing pubsub message: %v", err))
		return
	}

	mID := fmt.Sprintf("mid:%s", m.Message.MessageId)

	log.Printf("%s - message data: %s\n", mID, m.Message.Data)

	d, err := base64.StdEncoding.DecodeString(m.Message.Data)
	if err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error decoding message data: %v", err))
		return
	}

	log.Printf("%s - event data: %s\n", mID, d)

	var e event
	if err = json.Unmarshal(d, &e); err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error parsing event: %v", err))
		return
	}
	fmt.Printf("%s - event action: %s, digest: %s, tag: %s\n", mID, e.Action, e.Digest, e.Tag)

	if e.Action != actionInsert {
		writeMessage(w, http.StatusOK, fmt.Sprintf("unsupported event type: %s", e.Action))
		return
	}

	if m.Subscription == testSubscription {
		fmt.Printf("%s - skipping executing command during test", mID)
		writeMessage(w, http.StatusOK, "ok")
		return
	}

	cmd := exec.CommandContext(r.Context(), "/bin/bash",
		commandName, e.Digest, projectID, signKey, mID) //nolint:gosec
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		writeMessage(w, http.StatusInternalServerError, fmt.Sprintf("error executing command: %v", err))
		return
	}

	fmt.Printf("mid:%s - command output: %s\n", m.Message.MessageId, out)
	writeMessage(w, http.StatusOK, "ok")
}

func main() {
	http.HandleFunc("/", handler)

	if projectID == "" || signKey == "" {
		panic("either PROJECT_ID or SIGN_KEY env vars aren't set")
	}

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
