package handler

import (
	"strings"
	"testing"
)

func TestSeverityString(t *testing.T) {
	s := SeverityMedium
	if s.String() != severityParamNames[SeverityMedium] {
		t.Fatalf("wrong severity, expected: %s, got: %s",
			severityParamNames[SeverityMedium], s.String())
	}
}

func TestSeverityParsing(t *testing.T) {
	// bad
	if _, err := toScannerSeverityArg("bad"); err == nil {
		t.FailNow()
	}
	// good
	if _, err := toScannerSeverityArg(severityParamNames[SeverityMedium]); err != nil {
		t.Fatal(err)
	}
	// good uppercase
	if _, err := toScannerSeverityArg(strings.ToUpper(severityParamNames[SeverityMedium])); err != nil {
		t.Fatal(err)
	}
	// min
	if _, err := toScannerSeverityArg(severityParamNames[SeverityUnknown]); err != nil {
		t.Fatal(err)
	}
	// max
	if _, err := toScannerSeverityArg(severityParamNames[SeverityCritical]); err != nil {
		t.Fatal(err)
	}
	// incremental
	s, err := toScannerSeverityArg(severityParamNames[SeverityCritical])
	if err != nil && s != "CRITICAL" {
		t.Fatalf("expected CRITICAL, go: %s", s)
	}
	s, err = toScannerSeverityArg(severityParamNames[SeverityHigh])
	if err != nil && s != "HIGH,CRITICAL" {
		t.Fatalf("expected HIGH,CRITICAL, go: %s", s)
	}
	s, err = toScannerSeverityArg(severityParamNames[SeverityMedium])
	if err != nil && s != "MEDIUM,HIGH,CRITICAL" {
		t.Fatalf("expected MEDIUM,HIGH,CRITICAL, go: %s", s)
	}
	s, err = toScannerSeverityArg(severityParamNames[SeverityLow])
	if err != nil && s != "LOW,MEDIUM,HIGH,CRITICAL" {
		t.Fatalf("expected LOW,MEDIUM,HIGH,CRITICAL, go: %s", s)
	}
}
