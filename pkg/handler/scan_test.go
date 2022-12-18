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

	// missing max severity status
	checkStatus(t, r, h.ScanHandler, http.StatusBadRequest)
}

func TestVulnerabilityCheck(t *testing.T) {
	testFile := "../../tests/vuln.json"
	testDigest := "us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator@sha256:07739bc68262d61a3d352cb4ad9341bc7e8376bef0c22d31a53758a6f8f58359"

	if err := checkVulnerability(testFile, testDigest, SeverityCritical); err != nil {
		t.Fatal(err)
	}

	if err := checkVulnerability(testFile, testDigest, SeverityHigh); err == nil {
		t.Fatal("schema version error expected")
	}

	if err := checkVulnerability(testFile, "bad", SeverityHigh); err == nil {
		t.Fatal("digest error expected")
	}
}
