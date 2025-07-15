package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artemaris/loyalty/internal/app"
	"github.com/artemaris/loyalty/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("App run failed: %v", err)
	}
}
