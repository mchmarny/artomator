package handler

type DiscoReport struct {
	Created string         `json:"created"`
	Counts  *DiscoCounts   `json:"meta"`
	Results []*DiscoResult `json:"results"`
}

type DiscoCounts struct {
	Unknown  int `json:"unknown"`
	Low      int `json:"low"`
	Medium   int `json:"medium"`
	High     int `json:"high"`
	Critical int `json:"critical"`
}

type DiscoResult struct {
	Target          string   `json:"target"`
	Tags            []string `json:"tag"`
	Digests         []string `json:"digest"`
	Source          string   `json:"source"`
	Vulnerabilities map[string]*DiscoVulnerabilities
}

type DiscoVulnerabilities struct {
	ID       string `json:"id"`
	Pkg      string `json:"pkg"`
	URL      string `json:"url"`
	Severity string `json:"severity"`
	Updated  string `json:"updated"`
}
