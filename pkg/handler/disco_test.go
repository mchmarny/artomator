package handler

import (
	"context"
	"path"
	"testing"

	"github.com/mchmarny/artomator/pkg/metric"
	"github.com/stretchr/testify/assert"
)

func TestDiscoParser(t *testing.T) {
	rep := runDiscoParserTest(t, "../../tests/reports", "")
	assert.NotNil(t, rep)
}

func TestDiscoParserWithCVE(t *testing.T) {
	cve := "CVE-2020-8912"
	rep := runDiscoParserTest(t, "../../tests/reports", cve)
	assert.NotNil(t, rep)
	assert.NotNil(t, rep.Filter)
	assert.Equal(t, rep.Filter.CVE, cve)
	assert.Greater(t, len(rep.Filter.Services), 0)
}

func runDiscoParserTest(t *testing.T, dir, cve string) *DiscoReport {
	c := &metric.ConsoleCounter{}
	assert.NotNil(t, c)

	ctx := context.TODO()
	rec := newReporter(c, dir, cve)
	assert.NotNil(t, rec)

	rep, err := rec.create(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, rep)

	err = rec.recorder.Flush(ctx)
	assert.NoError(t, err)

	err = rec.close(ctx)
	assert.NoError(t, err)

	return rep
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
