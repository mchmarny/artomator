package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

const (
	actionInsert      = "INSERT"
	commandDefault    = "artomator"
	addressDefault    = ":8080"
	sigTagSuffix      = ".sig"
	attTagSuffix      = ".att"
	shutdownTimeout   = 3
	readWriteTimeout  = 1
	readHeaderTimeout = 3
	idleServerTimeout = 60
)

var (
	version = "v0.0.1-default"

	projectID = os.Getenv("PROJECT_ID")
	signKey   = os.Getenv("SIGN_KEY")
	redisIP   = os.Getenv("REDIS_IP")
	redisPort = os.Getenv("REDIS_PORT")

	commandName = commandDefault

	requestKey key
	client     *redis.Client
)

type key int

func main() {
	fmt.Printf("starting artomator server %s\n", version)
	if projectID == "" || signKey == "" {
		panic("either PROJECT_ID or SIGN_KEY env vars aren't set")
	}

	fmt.Printf("redis %s:%s\n", redisIP, redisPort)
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisIP, redisPort),
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	address := addressDefault
	if val, ok := os.LookupEnv("PORT"); ok {
		address = fmt.Sprintf(":%s", val)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	ctx := context.Background()
	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadTimeout:       readWriteTimeout * time.Second,
		WriteTimeout:      readWriteTimeout * time.Second,
		IdleTimeout:       idleServerTimeout * time.Second,
		ReadHeaderTimeout: readHeaderTimeout * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, requestKey, l.Addr().String())
			return ctx
		},
	}

	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	} else if err != nil {
		log.Printf("error listening for server: %s\n", err)
	}
}
