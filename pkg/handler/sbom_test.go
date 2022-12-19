package handler

import (
	"net/http"
	"testing"
)

func TestSBOMHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/sbom", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := getTestHandler(t)

	checkStatus(t, req, h.SBOMHandler, http.StatusBadRequest)

	q := req.URL.Query()
	q.Add("format", "spdx")
	q.Add("digest", "region.pkg.dev/project/artomator/artomator@sha256:123")
	req.URL.RawQuery = q.Encode()

	checkStatus(t, req, h.SBOMHandler, http.StatusOK)
}

func TestSHAParser(t *testing.T) {
	_, err := parseSHA("us-west1-docker.pkg.dev/test/test/tester:v1.2.3")
	if err == nil {
		t.Errorf("no error from image with only a tag")
	}
	_, err = parseSHA("us-west1-docker.pkg.dev/test/test/tester")
	if err == nil {
		t.Errorf("no error from image without tag")
	}

	t1, err := parseSHA("us-west1-docker.pkg.dev/test/test/tester@sha256:123")
	checkErr(t, err)
	if t1 != "123" {
		t.Errorf("failed to parse SHA from a valid registry URI: (got:%s, want:123)", t1)
	}
}
