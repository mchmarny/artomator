package handler

import "testing"

func TestDiscoParser(t *testing.T) {
	r, err := processReports("../../tests/reports")
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("nil report")
	}
}
