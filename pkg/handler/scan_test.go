package handler

import (
	"net/http"
	"testing"
)

func TestScanHandler(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "/scan", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := getTestHandler(t)

	checkStatus(t, r, h.ScanHandler, http.StatusBadRequest)

	q := r.URL.Query()
	q.Add("digest", "region.pkg.dev/project/artomator/artomator@sha256:123")
	r.URL.RawQuery = q.Encode()

	checkStatus(t, r, h.ScanHandler, http.StatusOK)

	q.Add("severity", "low")
	q.Add("scope", "all-layers")
	r.URL.RawQuery = q.Encode()

	checkStatus(t, r, h.ScanHandler, http.StatusOK)
}
