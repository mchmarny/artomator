package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mchmarny/artomator/pkg/cache"
	"github.com/mchmarny/artomator/pkg/cmd"
	"github.com/mchmarny/artomator/pkg/handler"
	"github.com/mchmarny/artomator/pkg/metric"
)

const (
	serviceName    = "artomator"
	addressDefault = ":8080"

	closeTimeout = 3
	readTimeout  = 10
	writeTimeout = 600
)

var (
	version = "v0.0.1-default"

	projectID  = os.Getenv("PROJECT_ID")
	signingKey = os.Getenv("SIGN_KEY")
	redisIP    = os.Getenv("REDIS_IP")
	redisPort  = os.Getenv("REDIS_PORT")
	bucketName = os.Getenv("GCS_BUCKET")
)

type key int

func main() {
	log.SetFlags(log.Lshortfile)
	log.Printf("starting %s server (%s)...\n", serviceName, version)

	if projectID == "" || signingKey == "" {
		log.Fatal("either PROJECT_ID or SIGN_KEY env var not defined")
	}

	if projectID == "" || signingKey == "" {
		redisIP = "127.0.0.1"
		redisPort = "6379"
	}

	ctx := context.Background()
	c, err := cache.NewPersistedCache(ctx, redisIP, redisPort)
	if err != nil {
		log.Fatalf("error while creating cache: %v", err)
	}

	m, err := metric.NewAPICounter(projectID)
	if err != nil {
		log.Fatalf("error while creating counter: %v", err)
	}

	if err := m.Count(ctx, metric.MakeMetricType("server/start"), 1, nil); err != nil {
		log.Printf("unable to write metrics: %v", err)
	}

	h, err := handler.NewHandler(bucketName, c, m,
		cmd.NewBashCommand(handler.CommandNameEvent, "event", projectID, signingKey),
		cmd.NewBashCommand(handler.CommandNameSBOM, "sbom", projectID, signingKey),
		cmd.NewBashCommand(handler.CommandNameVerify, "verify", projectID, signingKey),
	)
	if err != nil {
		log.Fatalf("error while creating event handler: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.HandlerDefault)
	mux.HandleFunc("/event", h.EventHandler)
	mux.HandleFunc("/sbom", h.SBOMHandler)
	mux.HandleFunc("/verify", h.VerifyHandler)

	address := addressDefault
	if val, ok := os.LookupEnv("PORT"); ok {
		address = fmt.Sprintf(":%s", val)
	}

	run(ctx, mux, address)
}

var contextKey key

func run(ctx context.Context, mux *http.ServeMux, address string) {
	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: readTimeout * time.Second,
		WriteTimeout:      writeTimeout * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			// adding server address to ctx handler functions receives
			return context.WithValue(ctx, contextKey, l.Addr().String())
		},
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("error listening for server: %v\n", err)
		}
	}()
	log.Print("server started")

	<-done
	log.Print("server stopped")

	downCtx, cancel := context.WithTimeout(ctx, closeTimeout*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(downCtx); err != nil {
		log.Fatalf("error shuting server down: %v", err)
	}
}
