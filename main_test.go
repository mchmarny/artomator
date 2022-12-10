package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	tagEvent = `{
		"action": "INSERT",
		"digest": "us-east1-docker.pkg.dev/my-project/my-repo/hello-world@sha256:6ec128e26cd5",
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
		Subscription: testSubscription,
		Message: pubsub.PubsubMessage{
			MessageId: fmt.Sprintf("id-%d", time.Now().UnixNano()),
			Data:      d,
		},
	}
}

func runTest(event string, expectedStatus int, t *testing.T) {
	b, err := json.Marshal(getPubSubMessage(event))
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != expectedStatus {
		t.Errorf("handler returned wrong status: (got %v want %v)", status, expectedStatus)
	}
}

func TestHandler(t *testing.T) {
	runTest(invalidEvent, http.StatusOK, t)
	runTest(tagEvent, http.StatusPartialContent, t)
	runTest(validEvent, http.StatusPartialContent, t)
	runTest(sigEvent, http.StatusOK, t)
	runTest(attEvent, http.StatusOK, t)
}

func TestSHAParser(t *testing.T) {
	t1 := getSHA("us-west1-docker.pkg.dev/test/test/tester@sha256:123")
	if t1 != "123" {
		t.Errorf("failed to properly parse SHA from registry URI: (got %s want 123)", t1)
	}
	t2 := getSHA("us-west1-docker.pkg.dev/test/test/tester:v1.2.3")
	if t2 != "v1.2.3" {
		t.Errorf("failed to properly parse label from registry URI: (got %s want v1.2.3)", t2)
	}
	t3 := getSHA("us-west1-docker.pkg.dev/test/test/tester")
	if t3 != "" {
		t.Errorf("failed to properly parse label from registry URI: (got %s want '')", t3)
	}
}
