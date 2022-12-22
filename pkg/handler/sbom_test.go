package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSBOMHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/sbom", nil)
	assert.NoError(t, err)

	h := getTestHandler(t)
	assert.NotNil(t, h)

	err = h.Validate(CommandNameEvent)
	assert.NoError(t, err)

	checkStatus(t, req, h.SBOMHandler, http.StatusBadRequest)

	q := req.URL.Query()
	q.Add("format", "spdx")
	req.URL.RawQuery = q.Encode()

	checkStatus(t, req, h.SBOMHandler, http.StatusBadRequest)
}

func TestSHAParser(t *testing.T) {
	_, err := parseSHA("us-west1-docker.pkg.dev/test/test/tester:v1.2.3")
	assert.Error(t, err)

	_, err = parseSHA("us-west1-docker.pkg.dev/test/test/tester")
	assert.Error(t, err)

	t1, err := parseSHA("us-west1-docker.pkg.dev/test/test/tester@sha256:123")
	assert.NoError(t, err)
	assert.Equal(t, "123", t1)
}
