package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloudflare-dns-rotator/config"
	"cloudflare-dns-rotator/rotator"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Cloudflare DNS Rotator starting...")
	log.Printf("Loading configuration from %s", *configPath)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("  Domain: %s", cfg.Domain)
	log.Printf("  Record: %s.%s", cfg.RecordName, cfg.Domain)
	log.Printf("  Targets: %v", cfg.Targets)
	log.Printf("  Interval: %s", cfg.Interval)
	log.Printf("  Proxied: %v", cfg.Proxied)
	log.Printf("  TTL: %d", cfg.TTL)

	// Create context that listens for shutdown signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Create and initialize the rotator
	r, err := rotator.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create rotator: %v", err)
	}

	if err := r.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize rotator: %v", err)
	}

	log.Printf("Starting rotation loop...")

	// Run the rotation loop
	if err := r.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Rotator stopped with error: %v", err)
	}

	log.Printf("Cloudflare DNS Rotator stopped")
}
