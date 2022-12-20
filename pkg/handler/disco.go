package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mchmarny/artomator/pkg/metric"
	"github.com/pkg/errors"
)

const (
	CommandNameDisco = "disco"

	severityKey = "severity"
)

func (h *Handler) DiscoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("preparing discovery...")

	if err := h.Validate(CommandNameDisco); err != nil {
		log.Fatalf("service not configured")
	}

	id := uuid.NewString()
	dir, err := makeFolder(id)
	if err != nil {
		writeError(w, errors.Wrapf(err, "error creating context from: %s", id))
		return
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			log.Printf("error deleting context: %s\n", dir)
		}
	}()

	if err := h.commands[CommandNameDisco].Run(r.Context(), dir); err != nil {
		writeError(w, errors.Wrap(err, "error validating"))
		return
	}

	d, err := processReports(dir)
	if err != nil {
		log.Printf("error validating attestation: %v", err)
		if err := h.counter.Count(r.Context(), metric.MakeMetricType("disco/failed"), 1, nil); err != nil {
			log.Printf("unable to write metrics: %v", err)
		}
		writeError(w, errors.New("error creating report"))
		return
	}

	h.recordDiscoMetrics(r.Context(), d.Counts)

	writeContent(w, d)
}

func (h *Handler) recordDiscoMetrics(ctx context.Context, counts *DiscoCounts) {
	items := make([]*metric.Record, 0)

	for k, v := range counts.TotalExposures {
		m := &metric.Record{
			MetricType:  metric.MakeMetricType("cve/total"),
			MetricValue: v,
			Labels: map[string]string{
				severityKey: k,
			},
		}
		items = append(items, m)
	}

	for k, v := range counts.ProjectExposures {
		m := &metric.Record{
			MetricType:  metric.MakeMetricType("cve/project"),
			MetricValue: v,
			Labels: map[string]string{
				severityKey: k,
			},
		}
		items = append(items, m)
	}

	for k, v := range counts.ServiceExposures {
		m := &metric.Record{
			MetricType:  metric.MakeMetricType("cve/service"),
			MetricValue: v,
			Labels: map[string]string{
				severityKey: k,
			},
		}
		items = append(items, m)
	}

	if err := h.counter.CountAll(ctx, items...); err != nil {
		log.Printf("unable to write count metrics: %v", err)
	}
}

func processReports(dir string) (*DiscoReport, error) {
	if dir == "" {
		return nil, errors.New("path required")
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading files from dir: %s", dir)
	}

	report := &DiscoReport{
		Created: time.Now().Format(time.RFC3339),
		Counts: &DiscoCounts{
			TotalExposures:   make(map[string]int64),
			ProjectExposures: make(map[string]int64),
			ServiceExposures: make(map[string]int64),
		},
		Results: make([]*DiscoResult, 0),
	}

	for _, file := range files {
		if err := fileToDiscoService(dir, file.Name(), report); err != nil {
			return nil, errors.Wrapf(err, "error parsing: %s/%s", dir, file.Name())
		}
	}

	return report, nil
}

func fileToDiscoService(dir, file string, rez *DiscoReport) error {
	if dir == "" {
		return errors.New("dir required")
	}
	if file == "" {
		return errors.New("file required")
	}
	if rez == nil {
		return errors.New("rez required")
	}

	f := path.Join(dir, file)
	b, err := os.ReadFile(f)
	if err != nil {
		return errors.Wrapf(err, "error reading file: %s", f)
	}

	var r ScanReport
	if err := json.Unmarshal(b, &r); err != nil {
		return errors.Wrapf(err, "error parsing scanned report: %s", f)
	}

	svcFullName := toServiceName(file)
	prjName, srvName, ok := toProjectService(file)

	for _, z := range r.Results {
		d := &DiscoResult{
			Artifact:        r.ArtifactName,
			Service:         svcFullName,
			Source:          z.Target,
			Digests:         r.Metadata.RepoDigests,
			Vulnerabilities: make(map[string]*DiscoVulnerabilities),
		}

		for _, v := range z.Vulnerabilities {
			d.Vulnerabilities[v.VulnerabilityID] = &DiscoVulnerabilities{
				ID:       v.VulnerabilityID,
				Pkg:      v.PkgName,
				Version:  v.InstalledVersion,
				URL:      v.PrimaryURL,
				Severity: v.Severity,
				Updated:  v.LastModifiedDate,
			}

			meter(rez.Counts.TotalExposures, v.Severity, "", "")

			// only if the name parsing was successful
			if ok {
				meter(rez.Counts.ServiceExposures, v.Severity, prjName, srvName)
			}
		}
		rez.Results = append(rez.Results, d)
	}

	return nil
}

func meter(m map[string]int64, sev, proj, srv string) {
	switch sev {
	case VulnCountLow:
		m[VulnCountLow]++
	case VulnCountMedium:
		m[VulnCountMedium]++
	case VulnCountHigh:
		m[VulnCountHigh]++
	case VulnCountCritical:
		m[VulnCountCritical]++
	default:
		m[VulnCountUnknown]++
	}

	if proj != "" && srv != "" {
		switch sev {
		case VulnCountLow:
			m[fmt.Sprintf("%s/%s", proj, VulnCountLow)]++
			m[fmt.Sprintf("%s/%s", srv, VulnCountLow)]++
		case VulnCountMedium:
			m[fmt.Sprintf("%s/%s", proj, VulnCountMedium)]++
			m[fmt.Sprintf("%s/%s", srv, VulnCountMedium)]++
		case VulnCountHigh:
			m[fmt.Sprintf("%s/%s", proj, VulnCountHigh)]++
			m[fmt.Sprintf("%s/%s", srv, VulnCountHigh)]++
		case VulnCountCritical:
			m[fmt.Sprintf("%s/%s", proj, VulnCountCritical)]++
			m[fmt.Sprintf("%s/%s", srv, VulnCountCritical)]++
		default:
			m[fmt.Sprintf("%s/%s", proj, VulnCountUnknown)]++
			m[fmt.Sprintf("%s/%s", srv, VulnCountUnknown)]++
		}
	}
}

const (
	fileNamePartDeliminator   = "---"
	fileNameExpectedPartCount = 3
)

// example: cloudy-demos.us-west1.artomator.unknown
func toProjectService(fileName string) (project string, service string, ok bool) {
	n := toServiceName(fileName)
	p := strings.Split(n, ".")
	if len(p) == fileNameExpectedPartCount {
		return p[0], p[2], true
	}
	log.Printf("unable to parse project/service from: %s", n)
	return "", "", false
}

func toServiceName(fileName string) string {
	if len(strings.Split(fileName, fileNamePartDeliminator)) != fileNameExpectedPartCount {
		return strings.ReplaceAll(fileName, ".json", "")
	}
	fileName = strings.ReplaceAll(fileName, ".json", "")
	return strings.ReplaceAll(fileName, fileNamePartDeliminator, ".")
}
