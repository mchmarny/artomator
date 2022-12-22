package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerificationHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/verify", nil)
	assert.NoError(t, err)

	h := getTestHandler(t)
	assert.NotNil(t, h)
	err = h.Validate(CommandNameEvent)
	assert.NoError(t, err)

	checkStatus(t, req, h.VerifyHandler, http.StatusBadRequest)

	q := req.URL.Query()
	q.Add("digest", "region.pkg.dev/project/artomator/artomator@sha256:123")
	req.URL.RawQuery = q.Encode()

	checkStatus(t, req, h.VerifyHandler, http.StatusBadRequest)
}

func TestPredicateParser(t *testing.T) {
	_, err := validateAttestation("../../tests/attestation.json")
	assert.NoError(t, err)

	_, err = validateAttestation("../../tests/empty.json")
	assert.Error(t, err)
}
