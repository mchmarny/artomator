package handler

import (
	"net/http"
	"testing"
)

func TestVerificationHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/verify", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := getTestHandler(t)
	if err = h.Validate(CommandNameEvent); err != nil {
		t.Fatal(err)
	}

	checkStatus(t, req, h.VerifyHandler, http.StatusBadRequest)

	q := req.URL.Query()
	q.Add("digest", "region.pkg.dev/project/artomator/artomator@sha256:123")
	req.URL.RawQuery = q.Encode()

	checkStatus(t, req, h.VerifyHandler, http.StatusBadRequest)
}

func TestPredicateParser(t *testing.T) {
	if _, err := validateAttestation("../../tests/attestation.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := validateAttestation("../../tests/empty.json"); err == nil {
		t.Fatal(err)
	}
}
