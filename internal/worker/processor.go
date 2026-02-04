package worker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Worker polls for RECEIVED transactions and processes them
type Worker struct {
	DB            *pgxpool.Pool
	PollInterval  time.Duration
	BatchSize     int
	stopChan      chan struct{}
}

// NewWorker creates a new transaction processing worker
func NewWorker(db *pgxpool.Pool) *Worker {
	return &Worker{
		DB:           db,
		PollInterval: 5 * time.Second,  // Poll every 5 seconds
		BatchSize:    10,                // Process up to 10 transactions per batch
		stopChan:     make(chan struct{}),
	}
}

// Start begins the worker loop
func (w *Worker) Start(ctx context.Context) {
	log.Println("🚀 Worker started - polling for RECEIVED transactions")
	
	ticker := time.NewTicker(w.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⏹️  Worker stopped - context cancelled")
			return
		case <-w.stopChan:
			log.Println("⏹️  Worker stopped - stop signal received")
			return
		case <-ticker.C:
			// Process a batch of transactions
			if err := w.processBatch(ctx); err != nil {
				log.Printf("❌ Error processing batch: %v", err)
			}
		}
	}
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	close(w.stopChan)
}

// processBatch finds and processes RECEIVED transactions
func (w *Worker) processBatch(ctx context.Context) error {
	// Find transactions with status RECEIVED
	query := `
		SELECT id, transaction_hash, chain_id
		FROM transactions
		WHERE status = 'RECEIVED'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := w.DB.Query(ctx, query, w.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to query RECEIVED transactions: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var (
			id    string
			hash  string
			chain int
		)

		if err := rows.Scan(&id, &hash, &chain); err != nil {
			log.Printf("❌ Failed to scan row: %v", err)
			continue
		}

		// Process this transaction
		if err := w.processTransaction(ctx, id, hash, chain); err != nil {
			log.Printf("❌ Failed to process transaction %s: %v", hash, err)
		} else {
			count++
		}
	}

	if count > 0 {
		log.Printf("✅ Processed %d transactions", count)
	}

	return rows.Err()
}

// processTransaction simulates fetching and normalizing a transaction
func (w *Worker) processTransaction(ctx context.Context, id, hash string, chainID int) error {
	log.Printf("📥 Processing transaction: %s (chain: %d)", hash, chainID)

	// Update status to FETCHING
	if err := w.updateStatus(ctx, id, "FETCHING", ""); err != nil {
		return fmt.Errorf("failed to update status to FETCHING: %w", err)
	}

	// Simulate fetching from blockchain (this would be a real RPC call in production)
	txData, err := w.simulateBlockchainFetch(hash, chainID)
	if err != nil {
		// Mark as ERROR if fetch failed
		w.updateStatus(ctx, id, "ERROR", err.Error())
		return err
	}

	// Normalize and store transaction data
	if err := w.normalizeAndStore(ctx, id, txData); err != nil {
		w.updateStatus(ctx, id, "ERROR", err.Error())
		return fmt.Errorf("failed to normalize transaction: %w", err)
	}

	// Update status to CONFIRMED
	if err := w.updateStatus(ctx, id, "CONFIRMED", ""); err != nil {
		return fmt.Errorf("failed to update status to CONFIRMED: %w", err)
	}

	log.Printf("✅ Transaction %s processed successfully", hash)
	return nil
}

// updateStatus updates the transaction status and logs the event
func (w *Worker) updateStatus(ctx context.Context, txID, newStatus, errorReason string) error {
	// Begin transaction
	tx, err := w.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get current status
	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM transactions WHERE id = $1`, txID).Scan(&currentStatus)
	if err != nil {
		return err
	}

	// Update transaction status
	updateQuery := `
		UPDATE transactions 
		SET status = $1, updated_at = now(), error_reason = $2
		WHERE id = $3
	`
	_, err = tx.Exec(ctx, updateQuery, newStatus, errorReason, txID)
	if err != nil {
		return err
	}

	// Log status change in ingestion_events
	eventQuery := `
		INSERT INTO ingestion_events (transaction_id, previous_status, new_status, reason)
		VALUES ($1, $2, $3, $4)
	`
	reason := fmt.Sprintf("Status changed by worker: %s → %s", currentStatus, newStatus)
	if errorReason != "" {
		reason = errorReason
	}
	
	_, err = tx.Exec(ctx, eventQuery, txID, currentStatus, newStatus, reason)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit(ctx)
}

// BlockchainTransaction represents simulated blockchain data
type BlockchainTransaction struct {
	Hash        string `json:"hash"`
	ChainID     int    `json:"chain_id"`
	FromAddress string `json:"from"`
	ToAddress   string `json:"to"`
	Value       string `json:"value"`
	BlockNumber int64  `json:"block_number"`
	GasUsed     int64  `json:"gas_used"`
	Status      string `json:"status"`
}

// simulateBlockchainFetch simulates fetching transaction data from blockchain RPC
// In production, this would call eth_getTransactionByHash or similar
func (w *Worker) simulateBlockchainFetch(hash string, chainID int) (*BlockchainTransaction, error) {
	// Simulate network delay
	time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)

	// Simulate occasional failures (5% chance)
	if rand.Float32() < 0.05 {
		return nil, fmt.Errorf("blockchain RPC error: transaction not found")
	}

	// Generate realistic-looking mock data
	tx := &BlockchainTransaction{
		Hash:        hash,
		ChainID:     chainID,
		FromAddress: w.generateAddress(),
		ToAddress:   w.generateAddress(),
		Value:       w.generateValue(),
		BlockNumber: 12345000 + rand.Int63n(1000000),
		GasUsed:     21000 + rand.Int63n(100000),
		Status:      "success",
	}

	return tx, nil
}

// normalizeAndStore updates the transaction with normalized data
func (w *Worker) normalizeAndStore(ctx context.Context, txID string, data *BlockchainTransaction) error {
	query := `
		UPDATE transactions
		SET 
			from_address = $1,
			to_address = $2,
			value = $3,
			block_number = $4,
			gas_used = $5,
			updated_at = now()
		WHERE id = $6
	`

	_, err := w.DB.Exec(ctx, query,
		data.FromAddress,
		data.ToAddress,
		data.Value,
		data.BlockNumber,
		data.GasUsed,
		txID,
	)

	return err
}

// Helper functions to generate mock data

func (w *Worker) generateAddress() string {
	const chars = "0123456789abcdef"
	addr := "0x"
	for i := 0; i < 40; i++ {
		addr += string(chars[rand.Intn(len(chars))])
	}
	return addr
}

func (w *Worker) generateValue() string {
	// Generate value between 0.001 and 100 ETH (in wei)
	base := rand.Int63n(100000000000000000)     // Up to 0.1 ETH
	value := 1000000000000000 + base             // Add 0.001 ETH minimum
	return fmt.Sprintf("%d", value)
}

// GetStats returns worker statistics (for monitoring)
func (w *Worker) GetStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	// Count transactions by status
	query := `
		SELECT status, COUNT(*) 
		FROM transactions 
		GROUP BY status
	`

	rows, err := w.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}

	return stats, rows.Err()
}
