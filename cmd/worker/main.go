package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Wuzu11517/TxnFlow/internal/config"
	"github.com/Wuzu11517/TxnFlow/internal/db"
	"github.com/Wuzu11517/TxnFlow/internal/worker"
)

func main() {
	log.Println("============================================================================")
	log.Println("TxnFlow Transaction Processor Worker")
	log.Println("============================================================================")

	// Load configuration
	cfg := config.Load()

	// Connect to database
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Printf("✅ Connected to database: %s", cfg.DatabaseURL)

	// Create worker
	w := worker.NewWorker(pool)
	log.Printf("⚙️  Worker configuration:")
	log.Printf("   - Poll interval: %v", w.PollInterval)
	log.Printf("   - Batch size: %d", w.BatchSize)

	// Print initial stats
	if stats, err := w.GetStats(ctx); err == nil {
		log.Println("📊 Current transaction status counts:")
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
	log.Println("\n🛑 Shutdown signal received, stopping worker...")

	// Stop worker gracefully
	w.Stop()
	cancel()

	// Print final stats
	log.Println("\n📊 Final transaction status counts:")
	if stats, err := w.GetStats(context.Background()); err == nil {
		for status, count := range stats {
			log.Printf("   - %s: %d", status, count)
		}
	}

	log.Println("============================================================================")
	log.Println("Worker stopped gracefully")
	log.Println("============================================================================")
}
