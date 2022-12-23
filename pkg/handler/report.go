package handler

import "time"

func newDiscoReport() *DiscoReport {
	return &DiscoReport{
		Created: time.Now().Format(time.RFC3339),
		Counts: &DiscoCounter{
			Totals:   make(map[string]int64),
			Projects: make(map[string]int64),
			Services: make(map[string]int64),
			Regions:  make(map[string]int64),
		},
		Results: make([]*DiscoResult, 0),
	}
}

type DiscoReport struct {
	Created string         `json:"created"`
	Filter  *CVEFilter     `json:"filter,omitempty"`
	Counts  *DiscoCounter  `json:"counts"`
	Results []*DiscoResult `json:"results"`
}

type CVEFilter struct {
	CVE      string          `json:"cve"`
	Severity string          `json:"severity"`
	URL      string          `json:"url"`
	Services map[string]bool `json:"affectedServices"`
}

type DiscoCounter struct {
	Totals   map[string]int64 `json:"total"`
	Projects map[string]int64 `json:"projects"`
	Services map[string]int64 `json:"services"`
	Regions  map[string]int64 `json:"regions"`
}

type VulnCounter struct {
	Critical int64 `json:"critical"`
	High     int64 `json:"high"`
	Medium   int64 `json:"medium"`
	Low      int64 `json:"low"`
	Unknown  int64 `json:"unknown"`
}

type DiscoResult struct {
	Artifact        string   `json:"artifact"`
	Digests         []string `json:"digests"`
	Service         string   `json:"service"`
	Source          string   `json:"source"`
	Vulnerabilities map[string]*DiscoVulnerabilities
}

type DiscoVulnerabilities struct {
	ID       string `json:"id"`
	Pkg      string `json:"pkg"`
	Version  string `json:"version"`
	URL      string `json:"url"`
	Severity string `json:"severity"`
	Updated  string `json:"updated"`
}

type ScanReport struct {
	ArtifactName string
	Metadata     struct {
		RepoTags    []string
		RepoDigests []string
	}
	Results []struct {
		Target          string
		Vulnerabilities []struct {
			VulnerabilityID  string
			PkgName          string
			InstalledVersion string
			PrimaryURL       string
			Severity         string
			LastModifiedDate string
		}
	}
}
