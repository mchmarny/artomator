package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mchmarny/artomator/pkg/cache"
)

func getTestHandler(t *testing.T) *EventHandler {
	a := []string{
		"-c",
		"echo",
		"test",
	}
	h, err := NewEventHandler(a, a, a, "", cache.NewInMemoryCache())
	if err != nil {
		t.Fatal(err)
	}
	if err = h.Validate(); err != nil {
		t.Fatal(err)
	}
	return h
}

func TestProcessHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/process", nil)
	if err != nil {
		t.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("format", "spdx")
	q.Add("digest", "region.pkg.dev/project/artomator/artomator@sha256:123")
	req.URL.RawQuery = q.Encode()

	r := httptest.NewRecorder()
	h := getTestHandler(t)
	handler := http.HandlerFunc(h.ProcessHandler)
	handler.ServeHTTP(r, req)
	if r.Code != http.StatusOK {
		t.Errorf("handler returned unexpected status (want:%d, got:%d)",
			http.StatusOK, r.Code)
	}
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
		t.Errorf("failed to parse SHA from a valid registry URI: (got:%s, want:123)", t1)
	}
	_, err = parseSHA("us-west1-docker.pkg.dev/test/test/tester:v1.2.3")
	if err == nil {
		t.Errorf("no error from image with only a tag")
	}
	_, err = parseSHA("us-west1-docker.pkg.dev/test/test/tester")
	if err == nil {
		t.Errorf("no error from image without tag")
	}
}
