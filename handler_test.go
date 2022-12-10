package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/api/pubsub/v1"
)

const (
	invalidEvent = `{
		"action": "DELETE",
		"tag": "us-east1-docker.pkg.dev/my-project/my-repo/hello-world:1.1"
	}`
	validEvent = `{
		"action": "INSERT",
		"digest": "us-east1-docker.pkg.dev/my-project/my-repo/hello-world@sha256:6ec128e26cd5"
	}`
	sigEvent = `{
		"action": "INSERT",
		"digest": "us-east1-docker.pkg.dev/my-project/my-repo/hello-world@sha256:6ec128e26cd5",
		"tag": "us-west1-docker.pkg.dev/cloudy-demos/artomator/tester:sha256-59d78.sig"
	}`
	attEvent = `{
		"action": "INSERT",
		"digest": "us-east1-docker.pkg.dev/my-project/my-repo/hello-world@sha256:6ec128e26cd5",
		"tag": "us-west1-docker.pkg.dev/cloudy-demos/artomator/tester:sha256-59d78.att"
	}`
)

func getPubSubMessage(content string) *pubsubMessage {
	d := base64.StdEncoding.EncodeToString([]byte(content))
	return &pubsubMessage{
		Subscription: "test",
		Message: pubsub.PubsubMessage{
			MessageId: fmt.Sprintf("id-%d", time.Now().UnixNano()),
			Data:      d,
		},
	}
}

func runTest(event string, expectedStatus int, t *testing.T) {
	commandName = "test"
	b, err := json.Marshal(getPubSubMessage(event))
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(handler)
	handler.ServeHTTP(r, req)
	if status := r.Code; status != expectedStatus {
		t.Errorf("handler returned wrong status: (got %v want %v)", status, expectedStatus)
	}

	d, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("error reading body %v", err)
	}
	t.Log(string(d))
}

func TestHandlerWithInvalidEvent(t *testing.T) {
	runTest(invalidEvent, http.StatusOK, t)
}
func TestHandlerWithValidEvent(t *testing.T) {
	runTest(validEvent, http.StatusOK, t)
}
func TestHandlerWithSignatureEvent(t *testing.T) {
	runTest(sigEvent, http.StatusOK, t)
}
func TestHandlerWithAttestationEvent(t *testing.T) {
	runTest(attEvent, http.StatusOK, t)
}
