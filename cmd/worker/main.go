package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Wuzu11517/TxnFlow/internal/blockchain"
	"github.com/Wuzu11517/TxnFlow/internal/config"
	"github.com/Wuzu11517/TxnFlow/internal/db"
	"github.com/Wuzu11517/TxnFlow/internal/worker"
)

func main() {
	log.Println("TxnFlow Transaction Processor Worker")

	cfg := config.Load()

	// Validate Infura API key
	if cfg.InfuraAPIKey == "" {
		log.Println("⚠️  WARNING: INFURA_API_KEY not set")
		log.Println("   Set environment variable INFURA_API_KEY to use real blockchain data")
		log.Println("   Get a free API key at https://infura.io")
		log.Fatal("Cannot start worker without INFURA_API_KEY")
	}

	// Connect to database
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Printf("Connected to database: %s", cfg.DatabaseURL)

	// Initialize chain registry
	chainRegistry := blockchain.NewChainRegistry(cfg.InfuraAPIKey)
	supportedChains := chainRegistry.GetSupportedChains()
	log.Printf("Supported chains: %v", supportedChains)

	// Create worker
	w := worker.NewWorker(pool, chainRegistry)
	log.Printf("Worker configuration:")
	log.Printf("   - Poll interval: %v", w.PollInterval)
	log.Printf("   - Batch size: %d", w.BatchSize)
	log.Printf("   - Blockchain: Real Ethereum data via Infura")

	// Print initial stats
	if stats, err := w.GetStats(ctx); err == nil {
		log.Println("Current transaction status counts:")
		for status, count := range stats {
			log.Printf("   - %s: %d", status, count)
		}
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start worker in background
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go w.Start(workerCtx)

	// Wait for shutdown signal
	<-sigChan
	log.Println("\nShutdown signal received, stopping worker...")

	// Stop worker gracefully
	w.Stop()
	cancel()

	// Print final stats
	log.Println("\nFinal transaction status counts:")
	if stats, err := w.GetStats(context.Background()); err == nil {
		for status, count := range stats {
			log.Printf("   - %s: %d", status, count)
		}
	}

	log.Println("Worker stopped gracefully")
}
