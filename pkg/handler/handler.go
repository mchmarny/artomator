package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mchmarny/artomator/pkg/cache"
	"github.com/mchmarny/artomator/pkg/cmd"
	"github.com/pkg/errors"
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

func NewHandler(bucket string, cache cache.Cache, commands ...*cmd.Command) (*Handler, error) {
	if cache == nil {
		return nil, errors.New("cache service not set")
	}

	h := &Handler{
		bucket:   bucket,
		cache:    cache,
		commands: make(map[string]*cmd.Command),
	}

	for _, c := range commands {
		h.commands[c.Kind] = c
	}

	return h, nil
}

type Handler struct {
	bucket   string
	commands map[string]*cmd.Command
	cache    cache.Cache
}

func (h *Handler) Validate(cmdName string) error {
	if h.cache == nil {
		return errors.New("cache service not set")
	}
	if h.commands == nil {
		return errors.New("commands not defined")
	}
	if _, ok := h.commands[cmdName]; !ok {
		return errors.Errorf("command %s not configured", cmdName)
	}
	return nil
}

func (h *Handler) HandlerDefault(w http.ResponseWriter, r *http.Request) {
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
