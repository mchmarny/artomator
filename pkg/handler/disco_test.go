package handler

import (
	"context"
	"path"
	"testing"

	"github.com/mchmarny/artomator/pkg/metric"
	"github.com/stretchr/testify/assert"
)

func TestDiscoParser(t *testing.T) {
	c := &metric.ConsoleCounter{}
	assert.NotNil(t, c)

	ctx := context.TODO()
	rec := newReporter(c, "../../tests/reports")
	assert.NotNil(t, rec)

	rep, err := rec.create(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, rep)

	err = rec.recorder.Flush(ctx)
	assert.NoError(t, err)

	err = rec.close(ctx)
	assert.NoError(t, err)
}

func TestDiscoServiceParser(t *testing.T) {
	f := "cloudy-demos---us-west1---artomator.json"
	d := "test"
	fi, ok := parseFileInfo(d, f)
	assert.True(t, ok)
	assert.NotNil(t, fi)
	assert.Equal(t, fi.path, path.Join(d, f))
	assert.Equal(t, fi.name, "cloudy-demos/us-west1/artomator")
	assert.Equal(t, fi.project, "cloudy-demos")
	assert.Equal(t, fi.region, "us-west1")
	assert.Equal(t, fi.service, "artomator")
}
