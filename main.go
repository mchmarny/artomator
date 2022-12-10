package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	redis "github.com/go-redis/redis/v8"
)

const (
	actionInsert     = "INSERT"
	commandDefault   = "artomator"
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

	commandName = commandDefault

	client *redis.Client
)

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
