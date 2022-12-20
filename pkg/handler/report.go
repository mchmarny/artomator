package handler

const (
	VulnCountUnknown  = "unknown"
	VulnCountLow      = "low"
	VulnCountMedium   = "medium"
	VulnCountHigh     = "high"
	VulnCountCritical = "critical"
)

type DiscoReport struct {
	Created   string           `json:"created"`
	Exposures map[string]int64 `json:"exposures"`
	Results   []*DiscoResult   `json:"results"`
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
