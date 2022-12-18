package handler

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	SeverityUnknown Severity = iota
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

var (
	severityParamNames = []string{
		"unknown",
		"low",
		"medium",
		"high",
		"critical",
	}
)

type Severity int

func (s Severity) String() string {
	if s < SeverityUnknown || s > SeverityCritical {
		return severityParamNames[SeverityUnknown]
	}
	return severityParamNames[s]
}

func (s Severity) Index() int {
	return int(s)
}

func (s Severity) IsEqualOrHigher(v string) bool {
	t := SeverityUnknown
	for i, s := range severityParamNames {
		if strings.TrimSpace(strings.ToLower(v)) == s {
			t = Severity(i)
		}
	}
	return t >= s
}

func toSeverity(v string) (Severity, error) {
	for i, s := range severityParamNames {
		if strings.TrimSpace(strings.ToLower(v)) == s {
			return Severity(i), nil
		}
	}
	return SeverityUnknown, errors.Errorf("invalid severity type: %s (options: %s)", v, strings.Join(severityParamNames, ","))
}

func toScannerSeverityArg(v string) (string, error) {
	for i, s := range severityParamNames {
		if strings.TrimSpace(strings.ToLower(v)) == s {
			r := severityParamNames[i:]
			return strings.ToUpper(strings.Join(r, ",")), nil
		}
	}
	return severityParamNames[SeverityUnknown],
		errors.Errorf("invalid severity type: %s (options: %s)", v, strings.Join(severityParamNames, ","))
}
