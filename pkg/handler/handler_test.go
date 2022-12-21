package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mchmarny/artomator/pkg/cache"
	"github.com/mchmarny/artomator/pkg/cmd"
	"github.com/mchmarny/artomator/pkg/metric"
)

func getTestHandler(t *testing.T) *Handler {
	testCmd := "echo"
	c := &metric.ConsoleCounter{}

	h, err := NewHandler("", cache.NewInMemoryCache(), c,
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

func TestRegistryInfo(t *testing.T) {
	runRegistryTest(t,
		"us-east1-docker.pkg.dev/my-project/my-repo/hello-world@sha256:6ec",
		"us-east1-docker.pkg.dev")
	runRegistryTest(t,
		"us-east1-docker.pkg.dev/my-project/hello-world:v0.1.2",
		"us-east1-docker.pkg.dev")

	runRegistryNameTest(t, "us-west1-docker.pkg.dev/image:v1.2.3", "image:v1.2.3")
	runRegistryNameTest(t, "us-west1-docker.pkg.dev/folder/image:v1.2.4", "folder")
	runRegistryNameTest(t, "us-west1-docker.pkg.dev/project/folder/image:v1.2.5", "folder")
	runRegistryNameTest(t, "us-west1-docker.pkg.dev/project/reg/folder/image:v1.2.6", "reg/folder")
}

func runRegistryTest(t *testing.T, uri, want string) {
	v, err := parseRegistry(uri)
	if err != nil {
		t.Fatal(err)
	}
	if v != want {
		t.Fatalf("expected: %s, got: %s", want, v)
	}
}

func runRegistryNameTest(t *testing.T, uri, want string) {
	v, err := parseRegistryName(uri)
	if err != nil {
		t.Fatal(err)
	}
	if v != want {
		t.Fatalf("expected: %s, got: %s", want, v)
	}
}
