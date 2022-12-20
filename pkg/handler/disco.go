package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/mchmarny/artomator/pkg/metric"
	"github.com/pkg/errors"
)

const (
	CommandNameDisco = "disco"
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

	if err := h.counter.Count(r.Context(), metric.MakeMetricType("disco/processed"), 1, nil); err != nil {
		log.Printf("unable to write metrics: %v", err)
	}

	writeContent(w, d)
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
		Counts:  &DiscoCounts{},
		Results: make([]*DiscoResult, 0),
	}

	for _, file := range files {
		if err := fileToDiscoService(dir, file.Name(), report); err != nil {
			return nil, errors.Wrapf(err, "error parsing: %s/%s", dir, file.Name())
		}
	}

	return report, nil
}

type scanReport struct {
	Metadata struct {
		RepoTags    []string
		RepoDigests []string
	}
	Results []struct {
		Target          string
		Vulnerabilities []struct {
			VulnerabilityID  string
			PkgName          string
			PrimaryURL       string
			Severity         string
			LastModifiedDate string
		}
	}
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

	var r scanReport
	if err := json.Unmarshal(b, &r); err != nil {
		return errors.Wrapf(err, "error parsing scanned report: %s", f)
	}

	for _, z := range r.Results {
		d := &DiscoResult{
			Target:          z.Target,
			Tags:            r.Metadata.RepoTags,
			Digests:         r.Metadata.RepoDigests,
			Source:          file,
			Vulnerabilities: make(map[string]*DiscoVulnerabilities),
		}

		for _, v := range z.Vulnerabilities {
			d.Vulnerabilities[v.VulnerabilityID] = &DiscoVulnerabilities{
				ID:       v.VulnerabilityID,
				Pkg:      v.PkgName,
				URL:      v.PrimaryURL,
				Severity: v.Severity,
				Updated:  v.LastModifiedDate,
			}
			switch v.Severity {
			case "LOW":
				rez.Counts.Low++
			case "MEDIUM":
				rez.Counts.Medium++
			case "HIGH":
				rez.Counts.High++
			case "CRITICAL":
				rez.Counts.Critical++
			default:
				rez.Counts.Unknown++
			}
		}
		rez.Results = append(rez.Results, d)
	}
	return nil
}
