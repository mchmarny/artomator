package main

import (
	"testing"
)

func TestSHAParser(t *testing.T) {
	t1 := getSHA("us-west1-docker.pkg.dev/test/test/tester@sha256:123")
	if t1 != "123" {
		t.Errorf("failed to properly parse SHA from registry URI: (got %s want 123)", t1)
	}
	t2 := getSHA("us-west1-docker.pkg.dev/test/test/tester:v1.2.3")
	if t2 != "v1.2.3" {
		t.Errorf("failed to properly parse label from registry URI: (got %s want v1.2.3)", t2)
	}
	t3 := getSHA("us-west1-docker.pkg.dev/test/test/tester")
	if t3 != "" {
		t.Errorf("failed to properly parse label from registry URI: (got %s want '')", t3)
	}
}
