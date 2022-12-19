package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mchmarny/artomator/pkg/cache"
)

const (
	imageDigestQueryParamName = "digest"
	successResponseMessage    = "ok"
)

type result struct {
	Status  string `json:"status"`
	Image   string `json:"image,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewEventHandler(eventArgs, verifyArgs, scanArgs, sbomArgs []string, bucket string, cache cache.Cache) (*EventHandler, error) {
	h := &EventHandler{
		bucketName:    bucket,
		cacheService:  cache,
		eventCmdArgs:  eventArgs,
		verifyCmdArgs: verifyArgs,
		scanCmdArgs:   scanArgs,
		sbomCmdArgs:   sbomArgs,
	}

	if err := h.Validate(); err != nil {
		return nil, err
	}
	return h, nil
}

type EventHandler struct {
	bucketName    string
	eventCmdArgs  []string
	verifyCmdArgs []string
	scanCmdArgs   []string
	sbomCmdArgs   []string
	cacheService  cache.Cache
}

// Validate ensures the services has been created in valid state.
func (h *EventHandler) Validate() error {
	if h.eventCmdArgs == nil {
		return errors.New("event command args not set")
	}
	if h.verifyCmdArgs == nil {
		return errors.New("verify command args not set")
	}
	if h.scanCmdArgs == nil {
		return errors.New("scan command args not set")
	}
	if h.sbomCmdArgs == nil {
		return errors.New("sbom command args not set")
	}
	if h.cacheService == nil {
		return errors.New("cache service is required")
	}
	return nil
}

func (h *EventHandler) HandlerDefault(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeMessage(w, "hello")
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	log.Println(err)
	writeContent(w, result{
		Status: http.StatusText(http.StatusBadRequest),
		Error:  err.Error(),
	})
}

func writeImageMessage(w http.ResponseWriter, digest, msg string) {
	w.WriteHeader(http.StatusOK)
	log.Println(msg)
	writeContent(w, result{
		Status:  http.StatusText(http.StatusOK),
		Image:   digest,
		Message: msg,
	})
}

func writeMessage(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusOK)
	log.Println(msg)
	writeContent(w, result{
		Status:  http.StatusText(http.StatusOK),
		Message: msg,
	})
}

func writeContent(w http.ResponseWriter, content any) {
	if err := json.NewEncoder(w).Encode(content); err != nil {
		log.Printf("error encoding: %v - %v", content, err)
	}
}
