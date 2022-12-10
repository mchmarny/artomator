package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	redis "github.com/go-redis/redis/v8"
	"google.golang.org/api/pubsub/v1"
)

const (
	actionInsert     = "INSERT"
	commandName      = "artomator"
	portDefault      = "8080"
	testSubscription = "test"
	sigTagSuffix     = ".sig"
	attTagSuffix     = ".att"
)

var (
	projectID = os.Getenv("PROJECT_ID")
	signKey   = os.Getenv("SIGN_KEY")
	redisIP   = os.Getenv("REDIS_IP")
	redisPort = os.Getenv("REDIS_PORT")

	client *redis.Client
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

type result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func writeMessage(w http.ResponseWriter, s int, m string) {
	w.WriteHeader(s)
	if err := json.NewEncoder(w).Encode(result{Status: s, Message: m}); err != nil {
		log.Printf("error encoding message: %s (%d) - %v", m, s, err)
	}
}

const expectedURIParts = 2

func getSHA(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) != expectedURIParts {
		return ""
	}
	return parts[1]
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Println("processing event...")

	var m pubsubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error parsing pubsub message: %v", err))
		return
	}

	d, err := base64.StdEncoding.DecodeString(m.Message.Data)
	if err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error decoding message data: %v", err))
		return
	}
	log.Printf("event data: %s\n", string(d))

	var e event
	if err = json.Unmarshal(d, &e); err != nil {
		writeMessage(w, http.StatusBadRequest, fmt.Sprintf("error parsing event: %v", err))
		return
	}

	imgSHA := getSHA(e.Digest)
	fmt.Printf("event action: %s, digest: %s, tag: %s, sha:%s\n", e.Action, e.Digest, e.Tag, imgSHA)

	if e.Action != actionInsert {
		writeMessage(w, http.StatusOK, fmt.Sprintf("unsupported event type: %s", e.Action))
		return
	}

	if strings.HasSuffix(e.Tag, sigTagSuffix) || strings.HasSuffix(e.Tag, attTagSuffix) {
		writeMessage(w, http.StatusOK, fmt.Sprintf("signature or attestation event: %s", e.Tag))
		return
	}

	if m.Subscription == testSubscription {
		fmt.Println("skipping executing command during test")
		writeMessage(w, http.StatusPartialContent, "ok")
		return
	}

	// run this after test exit above
	cachedDigest, err := client.Get(r.Context(), imgSHA).Result()
	if err == redis.Nil {
		err = client.Set(r.Context(), imgSHA, e.Digest, 0).Err()
		if err != nil {
			writeMessage(w, http.StatusInternalServerError,
				fmt.Sprintf("error setting key: %s - %v", imgSHA, err))
			return
		}
	} else if err != nil {
		writeMessage(w, http.StatusInternalServerError,
			fmt.Sprintf("error getting key: %s - %v", imgSHA, err))
		return
	} else {
		if cachedDigest != e.Digest {
			writeMessage(w, http.StatusInternalServerError,
				fmt.Sprintf("same SHA (%s) diff URI (cached:%s, request:%s)",
					imgSHA, cachedDigest, e.Digest))
			return
		}
	}

	cmd := exec.CommandContext(r.Context(), "/bin/bash", //nolint:gosec
		commandName, e.Digest, projectID, signKey)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		writeMessage(w, http.StatusInternalServerError, fmt.Sprintf("error executing command: %v", err))
		return
	}

	log.Printf("message %s done: %s\n", m.Message.MessageId, string(out))
	writeMessage(w, http.StatusOK, "ok")
}

func main() {
	http.HandleFunc("/", handler)

	if projectID == "" || signKey == "" {
		panic("either PROJECT_ID or SIGN_KEY env vars aren't set")
	}

	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisIP, redisPort),
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = portDefault
		fmt.Printf("using default port %s\n", port)
	}
	address := fmt.Sprintf(":%s", port)
	fmt.Printf("starting server %s\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		panic(err)
	}
}
