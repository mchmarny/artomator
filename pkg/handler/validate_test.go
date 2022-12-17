package handler

import (
	"net/http"
	"testing"
)

func TestValidationHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/validate", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := getTestHandler(t)

	checkStatus(t, req, h.ValidationHandler, http.StatusBadRequest)

	q := req.URL.Query()
	q.Add("digest", "region.pkg.dev/project/artomator/artomator@sha256:123")
	req.URL.RawQuery = q.Encode()

	checkStatus(t, req, h.ValidationHandler, http.StatusOK)

	q.Add("format", "spdx")
	req.URL.RawQuery = q.Encode()

	checkStatus(t, req, h.ValidationHandler, http.StatusOK)
}
