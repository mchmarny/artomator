package handler

import (
	"context"
	"path"
	"testing"

	"github.com/mchmarny/artomator/pkg/metric"
)

func TestDiscoParser(t *testing.T) {
	c := &metric.ConsoleCounter{}
	ctx := context.TODO()
	rec := newReporter(c, "../../tests/reports")
	rep, err := rec.create(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if rep == nil {
		t.Fatal("nil report")
	}
	if err := rec.recorder.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if err := rec.close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoServiceParser(t *testing.T) {
	f := "cloudy-demos---us-west1---artomator.json"
	d := "test"
	fi, ok := parseFileInfo(d, f)
	if !ok {
		t.Fatal("parse failed")
	}
	if fi.path != path.Join(d, f) {
		t.Fatal("join failed")
	}
	if fi.name != "cloudy-demos/us-west1/artomator" {
		t.Fatal("full name failed")
	}
	if fi.project != "cloudy-demos" {
		t.Fatal("project failed")
	}
	if fi.region != "us-west1" {
		t.Fatal("region")
	}
	if fi.service != "artomator" {
		t.Fatal("service")
	}
}
