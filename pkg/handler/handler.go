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

func NewEventHandler(processArgs, verifyArgs, scanArgs []string, bucket string, cache cache.Cache) (*EventHandler, error) {
	h := &EventHandler{
		bucketName:     bucket,
		cacheService:   cache,
		processCmdArgs: processArgs,
		verifyCmdArgs:  verifyArgs,
		scanCmdArgs:    scanArgs,
	}

	if err := h.Validate(); err != nil {
		return nil, err
	}
	return h, nil
}

type EventHandler struct {
	bucketName     string
	processCmdArgs []string
	verifyCmdArgs  []string
	scanCmdArgs    []string
	cacheService   cache.Cache
}

// Validate ensures the services has been created in valid state.
func (h *EventHandler) Validate() error {
	if h.processCmdArgs == nil {
		return errors.New("process command args not set")
	}
	if h.verifyCmdArgs == nil {
		return errors.New("verify command args not set")
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
		Status:  http.StatusText(http.StatusBadRequest),
		Message: err.Error(),
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
