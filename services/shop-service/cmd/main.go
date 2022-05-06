package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/karelrenaldi/storemono/libs/smarthttp"
	server "github.com/karelrenaldi/storemono/services/shop-service"
	"github.com/karelrenaldi/storemono/services/shop-service/internal/config"
	"github.com/karelrenaldi/storemono/services/shop-service/internal/constant"
)

const (
	shutdownTimeout = 10 * time.Second
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load env with err: %s\n", err)
	}

	fmt.Fprintf(os.Stderr, "before config.New()\n")

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config with err: %s\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "before newAppContext()\n")

	ctx, err := newAppContext(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "newAppContext error: %s\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "before server.NewServer()\n")

	server, err := server.NewServer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create the server instance with err: %s\n", err)
		return
	}

	go server.Listen()

	// listen for OS Signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// wait for OS signal
	<-signals

	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shutdown failed with err: %s\n", err)
	}
}

// newAppContext will create the global context and put in some useful resources, shared by all submodules.
func newAppContext(cfg *config.AppConfig) (ctx context.Context, err error) {
	ctx = context.Background()
	ctx = context.WithValue(ctx, constant.AppConfig, cfg)

	cli := &smarthttp.Client{
		Name: "smarthttp",
		Client: &http.Client{
			Timeout: cfg.HTTPClientTimeout(),
		},
		CircuitBreaker: smarthttp.CircuitBreaker{
			MaxConcurrentRequests: cfg.HTTPMaxConcurrency(),
		},
		Retries: &smarthttp.Retries{
			MaxAttempts: cfg.HTTPRetryAttempts(),
			BaseDelay:   cfg.HTTPRetryDelay(),
			MaxDelay:    cfg.HTTPRetryMaxDelay(),
		},
	}

	ctx = context.WithValue(ctx, constant.HTTPClient, cli)

	fmt.Fprintf(os.Stderr, "before database.New()\n")

	// db, err := database.New(cfg.DBConfig())
	// if err != nil {
	// 	return
	// }

	// ctx = context.WithValue(ctx, constant.DataService, db)

	return ctx, err
}
