package object

import (
	"testing"
)

func TestSHAParser(t *testing.T) {
	m, err := getFiles("test", ".")
	if err != nil {
		t.Fatal(err)
	}

	v, ok := m["save.go"]
	if !ok {
		t.Fatalf("didn't find the save file: %v", m)
	}
	if v != "test-save.go" {
		t.Fatalf("map key has unexpected value, want: test.go, got: %s", v)
	}
}
