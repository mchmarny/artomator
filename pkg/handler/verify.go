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
	verifyPredicateTypeParamName = "type"
	predicateTypeKey             = "predicateType"
)

func (h *EventHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("verifying request...")

	digest := r.URL.Query().Get(imageDigestQueryParamName)
	if digest == "" {
		writeError(w, errors.Errorf("verify %s parameter not set", imageDigestQueryParamName))
		return
	}

	predicateType := r.URL.Query().Get(verifyPredicateTypeParamName)
	if predicateType == "" {
		writeError(w, errors.Errorf("verify %s parameter not set", verifyPredicateTypeParamName))
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

	predicatePath := path.Join(dir, "predicate.json")
	if err := runCommand(r.Context(), append(h.verifyCmdArgs, digest, predicateType, predicatePath)); err != nil {
		writeError(w, errors.Wrap(err, "error executing verification"))
		return
	}

	att, err := validateAttestation(predicatePath)
	if err != nil {
		writeError(w, errors.Wrap(err, "image does not have a valid attestation"))
		return
	}

	writeContent(w, att)
}

func validateAttestation(path string) (map[string]interface{}, error) {
	if path == "" {
		return nil, errors.New("path required")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading attestation file: %s", path)
	}

	var att map[string]interface{}
	if err := json.Unmarshal(b, &att); err != nil {
		return nil, errors.Wrapf(err, "error parsing attestation file: %s", path)
	}

	if len(att) < 1 {
		return nil, errors.New("attestation not found")
	}

	attType, ok := att[predicateTypeKey]
	if !ok {
		return nil, errors.New("invalid attestation format")
	}

	log.Printf("attestation: %s, type: %d", path, attType)

	return att, nil
}
