package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const (
	validateFormatParamName = "format"

	validateFormatDefault = "spdx"
)

func (h *EventHandler) ValidationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("validating request...")

	if r.Method != http.MethodPost {
		writeError(w, errors.Errorf("method %s not supported, expected POST", r.Method))
		return
	}

	digest := r.URL.Query().Get(imageDigestQueryParamName)
	if digest == "" {
		writeError(w, errors.Errorf("validate %s parameter not set", imageDigestQueryParamName))
		return
	}

	sbomFmt := r.URL.Query().Get(validateFormatParamName)
	if sbomFmt == "" {
		sbomFmt = validateFormatDefault
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

	out, err := runCommand(r.Context(), append(h.validateCmdArgs, digest, sbomFmt, dir))
	if err != nil {
		writeError(w, errors.Wrap(err, "error executing validation"))
	}

	log.Printf("validation done: %s\n", string(out))

	writeMessage(w, "request validated")
}
