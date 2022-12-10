package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"google.golang.org/api/pubsub/v1"
)

type pubsubMessage struct {
	Message      pubsub.PubsubMessage `json:"message"`
	Subscription string               `json:"subscription"`
}

type event struct {
	Action string `json:"action"`
	Digest string `json:"digest"`
	Tag    string `json:"tag"`
}

func getEvent(r *http.Request) (*event, error) {
	var m pubsubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		return nil, errors.Wrap(err, "error parsing pubsub message")
	}

	d, err := base64.StdEncoding.DecodeString(m.Message.Data)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding message data")
	}

	var e event
	if err = json.Unmarshal(d, &e); err != nil {
		return nil, errors.Wrap(err, "error parsing event")
	}
	return &e, nil
}
