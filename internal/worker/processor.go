package worker

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Wuzu11517/TxnFlow/internal/blockchain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Worker polls for RECEIVED transactions and processes them
type Worker struct {
	DB            *pgxpool.Pool
	ChainRegistry *blockchain.ChainRegistry
	PollInterval  time.Duration
	BatchSize     int
	stopChan      chan struct{}
}

// NewWorker creates a new transaction processing worker
func NewWorker(db *pgxpool.Pool, chainRegistry *blockchain.ChainRegistry) *Worker {
	return &Worker{
		DB:            db,
		ChainRegistry: chainRegistry,
		PollInterval:  5 * time.Second, // Poll every 5 seconds
		BatchSize:     10,               // Process up to 10 transactions per batch
		stopChan:      make(chan struct{}),
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

// processTransaction fetches and normalizes a transaction from the blockchain
// NOW WITH EXPONENTIAL BACKOFF RETRY LOGIC!
func (w *Worker) processTransaction(ctx context.Context, id, hash string, chainID int) error {
	log.Printf("📥 Processing transaction: %s (chain: %d)", hash, chainID)

	// Update status to FETCHING
	if err := w.updateStatus(ctx, id, "FETCHING", ""); err != nil {
		return fmt.Errorf("failed to update status to FETCHING: %w", err)
	}

	// Fetch from blockchain with retry + exponential backoff
	var txData *BlockchainTransaction
	var err error

	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		txData, err = w.fetchFromBlockchain(ctx, hash, chainID)

		if err == nil {
			// Success!
			break
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			// Permanent error - don't retry
			log.Printf("❌ Permanent error for %s, not retrying: %v", hash, err)
			break
		}

		// Exponential backoff (only if we have more attempts)
		if attempt < maxRetries-1 {
			// Wait time: 2^attempt seconds (1s, 2s, 4s)
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("⚠️  Attempt %d/%d failed for %s, retrying in %v: %v", 
				attempt+1, maxRetries, hash, waitTime, err)
			time.Sleep(waitTime)
		}
	}

	if err != nil {
		// Mark as ERROR after all retries exhausted
		errorMsg := fmt.Sprintf("failed after %d retries: %v", maxRetries, err)
		w.updateStatus(ctx, id, "ERROR", errorMsg)
		return fmt.Errorf(errorMsg)
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

// isRetryableError determines if an error is worth retrying
func isRetryableError(err error) bool {
	errStr := strings.ToLower(err.Error())

	// Retry on transient errors
	retryablePatterns := []string{
		"timeout",
		"rate limit",
		"429",                // HTTP 429 Too Many Requests
		"connection refused",
		"temporary failure",
		"i/o timeout",
		"context deadline exceeded",
		"network error",
		"502", // Bad Gateway
		"503", // Service Unavailable
		"504", // Gateway Timeout
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Don't retry on permanent errors
	permanentPatterns := []string{
		"transaction not found",
		"invalid hash",
		"unsupported chain",
		"invalid transaction hash",
		"malformed",
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	// Default: don't retry unknown errors (conservative approach)
	// Change to 'return true' if you want to retry by default
	return false
}

// fetchFromBlockchain fetches real transaction data from blockchain RPC
func (w *Worker) fetchFromBlockchain(ctx context.Context, hash string, chainID int) (*BlockchainTransaction, error) {
	// Get chain configuration
	chainConfig, err := w.ChainRegistry.GetChain(chainID)
	if err != nil {
		return nil, fmt.Errorf("unsupported chain: %w", err)
	}

	// Create RPC client for this chain
	rpcClient := blockchain.NewRPCClient(chainConfig.RPCURL)

	// Fetch transaction
	ethTx, err := rpcClient.GetTransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}

	// Fetch receipt for gas usage and status
	receipt, err := rpcClient.GetTransactionReceipt(ctx, hash)
	if err != nil {
		// Receipt might not exist for pending transactions
		log.Printf("⚠️  No receipt found for %s (may be pending): %v", hash, err)
		// Continue without receipt data
	}

	// Convert to normalized format
	txData := &BlockchainTransaction{
		Hash:        ethTx.Hash,
		ChainID:     chainID,
		FromAddress: ethTx.From,
		ToAddress:   ethTx.To,
	}

	// Convert hex value to decimal string
	if ethTx.Value != "" {
		valueDecimal, err := blockchain.HexToDecimalString(ethTx.Value)
		if err != nil {
			log.Printf("⚠️  Failed to parse value: %v", err)
		} else {
			txData.Value = valueDecimal
		}
	}

	// Parse block number
	if ethTx.BlockNumber != "" {
		blockNum, err := blockchain.HexToInt64(ethTx.BlockNumber)
		if err != nil {
			log.Printf("⚠️  Failed to parse block number: %v", err)
		} else {
			txData.BlockNumber = blockNum
		}
	}

	// Parse gas used from receipt
	if receipt != nil && receipt.GasUsed != "" {
		gasUsed, err := blockchain.HexToInt64(receipt.GasUsed)
		if err != nil {
			log.Printf("⚠️  Failed to parse gas used: %v", err)
		} else {
			txData.GasUsed = gasUsed
		}

		// Parse status from receipt (0x1 = success, 0x0 = failed)
		if receipt.Status == "0x1" {
			txData.Status = "success"
		} else if receipt.Status == "0x0" {
			txData.Status = "failed"
		}
	}

	return txData, nil
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

// BlockchainTransaction represents normalized blockchain data
type BlockchainTransaction struct {
	Hash        string
	ChainID     int
	FromAddress string
	ToAddress   string
	Value       string
	BlockNumber int64
	GasUsed     int64
	Status      string
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