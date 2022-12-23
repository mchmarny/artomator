package pubsub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"google.golang.org/api/pubsub/v1"
)

// PubsubMessage is the PubSeb envelope for GCR event.
type PubsubMessage struct {
	Message      pubsub.PubsubMessage `json:"message"`
	Subscription string               `json:"subscription"`
}

// Event represents Artifact Registry event.
type Event struct {
	Action string `json:"action"`
	Digest string `json:"digest"`
	Tag    string `json:"tag"`
}

// ParseEvent parses the event from the request.
func ParseEvent(r *http.Request) (*Event, error) {
	var m PubsubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		return nil, errors.Wrap(err, "error parsing pubsub message")
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling pubsub message for debug")
	}
	fmt.Println(string(b))

	d, err := base64.StdEncoding.DecodeString(m.Message.Data)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding message data")
	}

	var e Event
	if err = json.Unmarshal(d, &e); err != nil {
		return nil, errors.Wrap(err, "error parsing event")
	}
	return &e, nil
}

func GetPubSubMessage(sub, content string) *PubsubMessage {
	return &PubsubMessage{
		Subscription: sub,
		Message: pubsub.PubsubMessage{
			MessageId: fmt.Sprintf("id-%d", time.Now().UnixNano()),
			Data:      base64.StdEncoding.EncodeToString([]byte(content)),
		},
	}
}
