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
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mchmarny/artomator/pkg/metric"
	"github.com/mchmarny/artomator/pkg/object"
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

	rec := newReporter(h.counter, dir)
	defer func() {
		if err = rec.close(r.Context()); err != nil {
			log.Printf("error closing recorder: %s\n", dir)
		}
	}()

	rep, err := rec.create(r.Context())
	if err != nil {
		log.Printf("error validating attestation: %v", err)
		if err := h.counter.Count(r.Context(), metric.MakeMetricType("disco/failed"), 1, nil); err != nil {
			log.Printf("unable to write metrics: %v", err)
		}
		writeError(w, errors.New("error creating report"))
		return
	}

	if h.bucket != "" {
		b, err := json.Marshal(rep)
		if err != nil {
			writeError(w, err)
			return
		}

		reportName := getDiscoReportName("cloud-run")
		if err := object.Put(r.Context(), h.bucket, reportName, b); err != nil {
			writeError(w, errors.Wrapf(err, "error writing content to: %s/%s",
				h.bucket, reportName))
			return
		}
	}

	writeContent(w, rep)
}

func getDiscoReportName(prefix string) string {
	return fmt.Sprintf("disco/%s/vuln-report-%s.json",
		prefix, time.Now().Format("2006-01-02"))
}

func newReporter(counter metric.Counter, dir string) *reporter {
	return &reporter{
		recorder: metric.NewRecorder(counter, nil),
		dir:      dir,
	}
}

type reporter struct {
	report   *DiscoReport
	recorder *metric.Recorder
	dir      string
	lock     sync.Mutex
}

func (r *reporter) close(ctx context.Context) error {
	if r.recorder == nil {
		return errors.New("recorder required")
	}
	if err := r.recorder.Flush(ctx); err != nil {
		return errors.Wrap(err, "error closing recorder")
	}
	return nil
}

func (r *reporter) create(ctx context.Context) (*DiscoReport, error) {
	if r.dir == "" {
		return nil, errors.New("path required")
	}
	if r.recorder == nil {
		return nil, errors.New("recorder required")
	}
	if r.report != nil {
		return nil, errors.New("reporter already initialized, create new one")
	}

	files, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading files from dir: %s", r.dir)
	}

	r.report = newDiscoReport()

	for _, file := range files {
		if err := r.processFile(ctx, file.Name()); err != nil {
			return nil, errors.Wrapf(err, "error parsing: %s/%s", r.dir, file.Name())
		}
	}

	return r.report, nil
}

func (r *reporter) processFile(ctx context.Context, file string) error {
	if r.dir == "" {
		return errors.New("path not set")
	}
	if r.recorder == nil {
		return errors.New("recorder not set")
	}
	if r.report == nil {
		return errors.New("report not created")
	}

	fi, ok := parseFileInfo(r.dir, file)
	b, err := os.ReadFile(fi.path)
	if err != nil {
		return errors.Wrapf(err, "error reading file: %s", fi.path)
	}

	var sr ScanReport
	if err := json.Unmarshal(b, &sr); err != nil {
		return errors.Wrapf(err, "error parsing scanned report: %+v", fi)
	}

	for _, z := range sr.Results {
		d := &DiscoResult{
			Artifact:        sr.ArtifactName,
			Service:         fi.name,
			Source:          z.Target,
			Digests:         sr.Metadata.RepoDigests,
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

			if err := r.recorder.Add(ctx, "cve", map[string]string{
				"project":  fi.project,
				"region":   fi.region,
				"service":  fi.service,
				"severity": strings.ToLower(v.Severity),
				"code":     v.VulnerabilityID,
			}); err != nil {
				log.Printf("error adding cve metric: %v", err)
			}

			r.lock.Lock()
			r.report.Counts.Totals[strings.ToLower(v.Severity)]++
			r.lock.Unlock()

			// only if the name parsing was successful
			if ok {
				r.lock.Lock()
				r.report.Counts.Projects[sevLabel(fi.project, v.Severity)]++
				r.report.Counts.Services[sevLabel(fi.service, v.Severity)]++
				r.report.Counts.Regions[sevLabel(fi.region, v.Severity)]++
				r.lock.Unlock()
			}
		}

		r.lock.Lock()
		r.report.Results = append(r.report.Results, d)
		r.lock.Unlock()
	}

	return nil
}

func sevLabel(pref, sev string) string {
	return fmt.Sprintf("%s/%s", pref, strings.ToLower(sev))
}

type fileInfo struct {
	// name - a/b/c
	name string
	// path - dir/file.json
	path string
	// project - cloudy-demos
	project string
	// region - artomator
	service string
	// region - us-west1
	region string
}

const (
	fileNamePartDeliminator    = "---"
	fileNameExpectedPartCount  = 3
	labelNameExpectedPartCount = 2
)

// example: cloudy-demos.us-west1.artomator
func parseFileInfo(dir, name string) (*fileInfo, bool) {
	cleanName := strings.ReplaceAll(name, ".json", "")
	fi := &fileInfo{
		path: path.Join(dir, name),
		name: strings.ReplaceAll(cleanName, fileNamePartDeliminator, "/"),
	}

	p := strings.Split(fi.name, "/")
	if len(p) != fileNameExpectedPartCount {
		log.Printf("invalid number of name parts, want: %d, got: %d",
			fileNameExpectedPartCount, len(p))
		return fi, false
	}

	fi.project = p[0]
	fi.region = p[1]
	fi.service = p[2]

	return fi, true
}
