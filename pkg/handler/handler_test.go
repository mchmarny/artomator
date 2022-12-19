package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mchmarny/artomator/pkg/cache"
	"github.com/mchmarny/artomator/pkg/cmd"
)

func getTestHandler(t *testing.T) *Handler {
	testCmd := "echo"
	h, err := NewHandler("", cache.NewInMemoryCache(),
		cmd.NewCommand(CommandNameEvent, testCmd),
		cmd.NewCommand(CommandNameSBOM, testCmd),
		cmd.NewCommand(CommandNameVerify, testCmd),
	)
	if err != nil {
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
