package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/pkg/errors"
)

const (
	maxSeverityParamName             = "max-vuln-severity"
	vulnerabilityReportSchemaVersion = 2
)

func (h *EventHandler) ScanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("processing scan request...")

	if r.Method != http.MethodPost {
		writeError(w, errors.Errorf("method %s not supported, expected POST", r.Method))
		return
	}

	digest := r.URL.Query().Get(imageDigestQueryParamName)
	if digest == "" {
		writeError(w, errors.Errorf("verify %s parameter not set", imageDigestQueryParamName))
		return
	}

	sevStr := r.URL.Query().Get(maxSeverityParamName)
	maxSeverity, err := toSeverity(sevStr)
	if err != nil {
		writeError(w, errors.Wrap(err, "invalid severity"))
		return
	}

	severityArg, err := toScannerSeverityArg(maxSeverity.String())
	if err != nil {
		writeError(w, errors.Wrap(err, "error parsing severity"))
		return
	}

	sha, err := parseSHA(digest)
	if err != nil {
		writeError(w, errors.Wrap(err, "error parsing process event sha"))
		return
	}

	dir, err := makeFolder(sha)
	if err != nil {
		writeError(w, errors.Wrapf(err, "error creating context from sha: %s", sha))
		return
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			log.Printf("error deleting context: %s\n", dir)
		}
	}()

	reportPath := path.Join(dir, "report.json")
	scanCmdArgs := append(h.scanCmdArgs, digest, severityArg, reportPath)
	if err := runCommand(r.Context(), scanCmdArgs); err != nil {
		writeError(w, errors.Wrap(err, "error executing validation"))
		return
	}

	if err := checkVulnerability(reportPath, digest, maxSeverity); err != nil {
		writeError(w, errors.Wrap(err, "image validation failed"))
		return
	}

	writeImageMessage(w, digest, "image valid")
}

type vulnerabilityReport struct {
	SchemaVersion int
	ArtifactName  string
	Results       []struct {
		Target          string
		Vulnerabilities []struct {
			VulnerabilityID string
			Severity        string
		}
	}
}

func checkVulnerability(path, digest string, max Severity) error {
	if path == "" || digest == "" {
		return errors.New("path and digest required")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "error reading vulnerability file: %s", path)
	}

	var r vulnerabilityReport
	if err := json.Unmarshal(b, &r); err != nil {
		return errors.Wrapf(err, "error parsing vulnerability report: %s", path)
	}

	if r.SchemaVersion != vulnerabilityReportSchemaVersion {
		return errors.Errorf("invalid report schema version, expected: %d, go: %d)", vulnerabilityReportSchemaVersion, r.SchemaVersion)
	}

	if r.ArtifactName != digest {
		return errors.Errorf("report not for correct digest, expected: %s, go: %s)", digest, r.ArtifactName)
	}

	if len(r.Results) == 0 {
		return nil
	}

	for _, z := range r.Results {
		for _, v := range z.Vulnerabilities {
			if max.IsEqualOrHigher(v.Severity) {
				return errors.Errorf("report contains severity equal or higher to the provided maximum threshold: %s (image: %s contains %s in %s)", max, digest, v.VulnerabilityID, z.Target)
			}
		}
	}

	return nil
}
