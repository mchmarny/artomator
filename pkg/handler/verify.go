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
	verifyFormatParamName = "format"
	verifyFormatDefault   = "spdx"

	expectedAttestationPayloadType = "application/vnd.in-toto+json"
)

func (h *EventHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("verifying request...")

	if r.Method != http.MethodPost {
		writeError(w, errors.Errorf("method %s not supported, expected POST", r.Method))
		return
	}

	digest := r.URL.Query().Get(imageDigestQueryParamName)
	if digest == "" {
		writeError(w, errors.Errorf("verify %s parameter not set", imageDigestQueryParamName))
		return
	}

	sbomFmt := r.URL.Query().Get(verifyFormatParamName)
	if sbomFmt == "" {
		sbomFmt = verifyFormatDefault
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

	sbomPath := path.Join(dir, "sbom.json")
	if err := runCommand(r.Context(), append(h.verifyCmdArgs, digest, sbomFmt, sbomPath)); err != nil {
		writeError(w, errors.Wrap(err, "error executing verification"))
		return
	}

	if err := validateAttestation(sbomPath); err != nil {
		writeError(w, errors.Wrap(err, "image does not have a valid attestation"))
		return
	}

	writeImageMessage(w, digest, "image verified")
}

type attestation struct {
	PayloadType string `json:"payloadType"`
	Signatures  []struct {
		Signature string `json:"sig"`
	} `json:"signatures"`
}

func validateAttestation(path string) error {
	if path == "" {
		return errors.New("path required")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "error reading attestation file: %s", path)
	}

	var att attestation
	if err := json.Unmarshal(b, &att); err != nil {
		return errors.Wrapf(err, "error parsing attestation file: %s", path)
	}

	if att.PayloadType != expectedAttestationPayloadType {
		return errors.Errorf("invalid attestation type (want:%s, got:%s)", expectedAttestationPayloadType, att.PayloadType)
	}

	if len(att.Signatures) < 1 {
		return errors.New("attestation missing signatures")
	}

	return nil
}
