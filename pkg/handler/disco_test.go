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

func TestDiscoServiceParser(t *testing.T) {
	if s := toServiceName("cloudy-demos---us-west1---artomator.json"); s != "cloudy-demos/us-west1/artomator" {
		t.Fatalf("expected: cloudy-demos/us-west1/artomator, got: %s", s)
	}
	if s := toServiceName("cloudy-demos--us-west1--artomator.json"); s != "cloudy-demos--us-west1--artomator" {
		t.Fatalf("expected: cloudy-demos--us-west1--artomator, got: %s", s)
	}
}
