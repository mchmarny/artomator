package handler

type CVEFilter struct {
	CVE      string          `json:"cve"`
	Severity string          `json:"severity"`
	URL      string          `json:"url"`
	Services map[string]bool `json:"affectedServices"`
}

type VulnCounter struct {
	Critical int64 `json:"critical"`
	High     int64 `json:"high"`
	Medium   int64 `json:"medium"`
	Low      int64 `json:"low"`
	Unknown  int64 `json:"unknown"`
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
