package handler

const (
	VulnCountUnknown  = "UNKNOWN"
	VulnCountLow      = "LOW"
	VulnCountMedium   = "MEDIUM"
	VulnCountHigh     = "HIGH"
	VulnCountCritical = "CRITICAL"
)

type DiscoReport struct {
	Created string         `json:"created"`
	Counts  *DiscoCounts   `json:"counts"`
	Results []*DiscoResult `json:"results"`
}

type DiscoCounts struct {
	TotalExposures   map[string]int64 `json:"totalExposures"`
	ProjectExposures map[string]int64 `json:"projectExposures"`
	ServiceExposures map[string]int64 `json:"serviceExposures"`
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
