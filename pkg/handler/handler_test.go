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

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func checkStatus(t *testing.T, req *http.Request, f func(http.ResponseWriter, *http.Request), status int) {
	r := httptest.NewRecorder()
	handler := http.HandlerFunc(f)
	handler.ServeHTTP(r, req)
	if r.Code != status {
		t.Errorf("handler returned unexpected status (want:%d, got:%d)", status, r.Code)
	}
}
