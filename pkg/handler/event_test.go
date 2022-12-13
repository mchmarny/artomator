package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mchmarny/artomator/pkg/cache"
	"github.com/mchmarny/artomator/pkg/pubsub"
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

func runTest(event string, t *testing.T) {
	a := []string{
		"-c",
		"echo",
		"test",
	}
	h, err := NewEventHandler(a, "", cache.NewInMemoryCache())
	if err != nil {
		t.Fatal(err)
	}
	if err = h.validate(); err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(pubsub.GetPubSubMessage("test", event))
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRecorder()
	handler := http.HandlerFunc(h.HandleEvent)
	handler.ServeHTTP(r, req)
	if status := r.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status: %d", status)
	}

	d, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("error reading body %v", err)
	}
	t.Log(string(d))
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestSHAParser(t *testing.T) {
	t1, err := parseSHA("us-west1-docker.pkg.dev/test/test/tester@sha256:123")
	checkErr(t, err)
	if t1 != "123" {
		t.Errorf("failed to properly parse SHA from registry URI: (got %s want 123)", t1)
	}
	t2, err := parseSHA("us-west1-docker.pkg.dev/test/test/tester:v1.2.3")
	checkErr(t, err)
	if t2 != "v1.2.3" {
		t.Errorf("failed to properly parse label from registry URI: (got %s want v1.2.3)", t2)
	}
	t3, err := parseSHA("us-west1-docker.pkg.dev/test/test/tester")
	checkErr(t, err)
	if t3 != "" {
		t.Errorf("failed to properly parse label from registry URI: (got %s want '')", t3)
	}
}

func TestHandlerWithInvalidEvent(t *testing.T) {
	runTest(invalidEvent, t)
}
func TestHandlerWithValidEvent(t *testing.T) {
	runTest(validEvent, t)
}
func TestHandlerWithSignatureEvent(t *testing.T) {
	runTest(sigEvent, t)
}
func TestHandlerWithAttestationEvent(t *testing.T) {
	runTest(attEvent, t)
}
