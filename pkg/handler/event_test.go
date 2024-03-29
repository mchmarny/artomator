package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mchmarny/artomator/pkg/pubsub"
	"github.com/stretchr/testify/assert"
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
		"tag": "us-west1-docker.pkg.dev/s3cme1/artomator/tester:sha256-59d78.sig"
	}`
	attEvent = `{
		"action": "INSERT",
		"digest": "us-east1-docker.pkg.dev/my-project/my-repo/hello-world@sha256:6ec128e26cd5",
		"tag": "us-west1-docker.pkg.dev/s3cme1/artomator/tester:sha256-59d78.att"
	}`
)

func TestEventHandlerWithInvalidEvent(t *testing.T) {
	runEventTest(t, invalidEvent)
}
func TestEventHandlerWithValidEvent(t *testing.T) {
	runEventTest(t, validEvent)
}
func TestEventHandlerWithSignatureEvent(t *testing.T) {
	runEventTest(t, sigEvent)
}
func TestEventHandlerWithAttestationEvent(t *testing.T) {
	runEventTest(t, attEvent)
}

func runEventTest(t *testing.T, event string) {
	b, err := json.Marshal(pubsub.GetPubSubMessage("test", event))
	assert.NoError(t, err)
	assert.NotNil(t, b)

	req, err := http.NewRequest(http.MethodPost, "/event", bytes.NewBuffer(b))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	h := getTestHandler(t)
	assert.NotNil(t, h)

	err = h.Validate(CommandNameEvent)
	assert.NoError(t, err)

	checkStatus(t, req, h.EventHandler, http.StatusOK)
}
