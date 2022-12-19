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

	"github.com/mchmarny/artomator/pkg/metric"
	"github.com/pkg/errors"
)

const (
	expectedURIParts    = 2
	actionInsert        = "INSERT"
	sigTagSuffix        = ".sig"
	attTagSuffix        = ".att"
	sbomFormatParamName = "format"
	spdxVersionKey      = "spdxVersion"

	CommandNameSBOM = "sbom"
)

func (h *Handler) SBOMHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("processing event...")

	if err := h.Validate(CommandNameSBOM); err != nil {
		log.Fatalf("service not configured")
	}

	digest := r.URL.Query().Get(imageDigestQueryParamName)
	if digest == "" {
		writeError(w, errors.Errorf("process %s parameter not set", imageDigestQueryParamName))
		return
	}

	ri, err := getRegInfo(digest)
	if err != nil {
		writeError(w, err)
		return
	}

	m, err := h.processSBOM(r.Context(), digest)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.counter.Count(r.Context(), metric.MakeMetricType("sbom/processed"), 1, ri); err != nil {
		log.Printf("unable to write metrics: %v", err)
	}

	writeContent(w, m)
}

func (h *Handler) processSBOM(ctx context.Context, digest string) (map[string]interface{}, error) {
	log.Printf("processing digest: %s", digest)

	sha, err := parseSHA(digest)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing process event sha")
	}

	dir, err := makeFolder(sha)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating context from sha: %s", sha)
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			log.Printf("error deleting context: %s\n", dir)
		}
	}()

	sbomPath := path.Join(dir, "sbom.json")
	if err := h.commands[CommandNameSBOM].Run(ctx, digest, sbomPath); err != nil {
		return nil, errors.Wrap(err, "error executing command")
	}

	sbom, err := validateSBOM(sbomPath)
	if err != nil {
		return nil, errors.Wrap(err, "image does not have a valid attestation")
	}

	return sbom, nil
}

func validateSBOM(path string) (map[string]interface{}, error) {
	if path == "" {
		return nil, errors.New("path required")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading sbom file: %s", path)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Wrapf(err, "error parsing sbom file: %s", path)
	}

	if len(m) < 1 {
		return nil, errors.New("sbom not found")
	}

	sbomVersion, ok := m[spdxVersionKey]
	if !ok {
		return nil, errors.New("invalid sbom format")
	}

	log.Printf("sbom: %s, version: %d", path, sbomVersion)

	return m, nil
}

// parseSHA ensures that the image URI is actually a digest
// shouldn't process based on labels
// example: "us-west1-docker.pkg.dev/test/test/tester@sha256:123"
func parseSHA(uri string) (string, error) {
	parts := strings.Split(uri, "@")
	if len(parts) != expectedURIParts {
		return "", errors.Errorf("unable to parse digest (@) from %s", uri)
	}

	parts = strings.Split(parts[1], ":")
	if len(parts) != expectedURIParts {
		return "", errors.Errorf("unable to parse SHA (:) from %s", uri)
	}
	return parts[1], nil
}

func makeFolder(sha string) (string, error) {
	p := fmt.Sprintf("./%s", sha)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(p, os.ModePerm); err != nil {
			return "", errors.Wrap(err, "error creating folder")
		}
	}
	return p, nil
}
