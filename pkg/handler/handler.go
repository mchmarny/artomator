package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mchmarny/artomator/pkg/cache"
	"github.com/mchmarny/artomator/pkg/cmd"
	"github.com/mchmarny/artomator/pkg/metric"
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

func NewHandler(bucket string, cache cache.Cache, counter metric.Counter, commands ...*cmd.Command) (*Handler, error) {
	if cache == nil {
		return nil, errors.New("cache service not set")
	}

	h := &Handler{
		bucket:   bucket,
		cache:    cache,
		counter:  counter,
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
	counter  metric.Counter
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

func parseRegistry(uri string) (string, error) {
	parts := strings.Split(uri, "/")
	if len(parts) < expectedURIParts {
		return "", errors.Errorf("unable to parse registry from %s", uri)
	}

	return parts[0], nil
}

const (
	regURIFormat2 = 2
	regURIFormat3 = 3
	regURIFormat4 = 4
	regURIFormat5 = 5
)

func parseRegistryName(uri string) (string, error) {
	parts := strings.Split(uri, "/")

	switch len(parts) {
	case regURIFormat2: // us-west1-docker.pkg.dev/image:v1.2.3
	case regURIFormat3: // us-west1-docker.pkg.dev/folder/image:v1.2.3
		return parts[1], nil
	case regURIFormat4: // us-west1-docker.pkg.dev/reg/folder/image:v1.2.3
		return fmt.Sprintf("%s/%s", parts[2], parts[3]), nil
	case regURIFormat5: // us-west1-docker.pkg.dev/project/reg/folder/image:v1.2.3
		return fmt.Sprintf("%s/%s", parts[3], parts[4]), nil
	}

	return "", errors.Errorf("unable to parse registry name from %s", uri)
}

func getRegInfo(uri string) (map[string]string, error) {
	r, err := parseRegistry(uri)
	if err != nil {
		return nil, err
	}
	n, err := parseRegistryName(uri)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"registry": r,
		"artifact": n,
	}, nil
}
