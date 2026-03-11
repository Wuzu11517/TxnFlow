# Background Worker: Transaction Processor

## Overview

The transaction processor worker is a background service that:
1. Polls for transactions with status `RECEIVED`
2. Simulates fetching transaction data from blockchain
3. Normalizes and stores the data
4. Updates transaction status through state transitions

This demonstrates production-grade async processing patterns.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  API Server (cmd/api)                                       │
│  - POST /transactions → Creates with status='RECEIVED'      │
│  - GET /transactions/{hash} → Returns current state         │
│  - GET /stats → Shows processing statistics                 │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │ Database
                     │ (shared)
                     │
┌────────────────────┴────────────────────────────────────────┐
│  Worker Process (cmd/worker)                                │
│  - Polls every 5 seconds                                    │
│  - Processes up to 10 transactions per batch                │
│  - Updates status: RECEIVED → FETCHING → CONFIRMED         │
└─────────────────────────────────────────────────────────────┘
```

## State Machine

Transactions flow through these status states:

```
RECEIVED
   ↓ (worker picks up)
FETCHING
   ↓ (simulated blockchain fetch)
CONFIRMED ✅
   
   OR
   
ERROR ❌ (if fetch fails)
```

**Status Definitions**:
- `RECEIVED`: Transaction hash registered, waiting for processing
- `FETCHING`: Worker is fetching data from blockchain
- `CONFIRMED`: Successfully processed with full data
- `ERROR`: Failed to fetch or process (with error_reason)

## How It Works

### 1. Worker Polling

Every 5 seconds, the worker:
```sql
SELECT id, transaction_hash, chain_id
FROM transactions
WHERE status = 'RECEIVED'
ORDER BY created_at ASC
LIMIT 10
```

Processes oldest transactions first (FIFO).

### 2. Status Update to FETCHING

```sql
UPDATE transactions 
SET status = 'FETCHING', updated_at = now()
WHERE id = $1
```

Also logs event:
```sql
INSERT INTO ingestion_events (transaction_id, previous_status, new_status, reason)
VALUES ($1, 'RECEIVED', 'FETCHING', 'Status changed by worker...')
```

### 3. Simulated Blockchain Fetch

In production, this would be:
```go
// Real implementation
rpcClient.eth_getTransactionByHash(hash)
```

For demo purposes, we simulate:
```go
func simulateBlockchainFetch(hash string, chainID int) (*BlockchainTransaction, error) {
    // Simulate 100-500ms network delay
    time.Sleep(100-500ms)
    
    // Simulate 5% failure rate
    if rand.Float32() < 0.05 {
        return nil, errors.New("blockchain RPC error")
    }
    
    // Generate realistic mock data
    return &BlockchainTransaction{
        FromAddress: "0x...",
        ToAddress:   "0x...",
        Value:       "1000000000000000000",  // 1 ETH in wei
        BlockNumber: 12345678,
        GasUsed:     21000,
    }, nil
}
```

### 4. Normalize and Store

Extract fields and update database:
```sql
UPDATE transactions
SET 
    from_address = $1,
    to_address = $2,
    value = $3,
    block_number = $4,
    gas_used = $5,
    updated_at = now()
WHERE id = $6
```

### 5. Final Status Update

```sql
UPDATE transactions 
SET status = 'CONFIRMED', updated_at = now()
WHERE id = $1
```

## Running the Worker

### Start the Worker

```bash
# Using default DATABASE_URL from environment
go run cmd/worker/main.go

# Or specify database URL
DATABASE_URL=postgres://user:pass@localhost/txnflow go run cmd/worker/main.go
```

### Worker Output

```
============================================================================
TxnFlow Transaction Processor Worker
============================================================================
✅ Connected to database: postgres://localhost/txnflow
⚙️  Worker configuration:
   - Poll interval: 5s
   - Batch size: 10
📊 Current transaction status counts:
   - RECEIVED: 5
   - CONFIRMED: 23
🚀 Worker started - polling for RECEIVED transactions

📥 Processing transaction: 0xabc123... (chain: 1)
✅ Transaction 0xabc123... processed successfully
📥 Processing transaction: 0xdef456... (chain: 1)
✅ Transaction 0xdef456... processed successfully
✅ Processed 2 transactions
```

### Graceful Shutdown

Press `Ctrl+C`:
```
^C
🛑 Shutdown signal received, stopping worker...
⏹️  Worker stopped - stop signal received

📊 Final transaction status counts:
   - RECEIVED: 0
   - CONFIRMED: 30
============================================================================
Worker stopped gracefully
============================================================================
```

## Testing the Full Flow

### 1. Start API Server

```bash
go run cmd/api/main.go
```

### 2. Start Worker (in separate terminal)

```bash
go run cmd/worker/main.go
```

### 3. Create Transaction

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_hash": "0xtest123",
    "chain_id": 1
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transaction_hash": "0xtest123",
  "chain_id": 1,
  "status": "RECEIVED",
  "created_at": "2025-01-30T12:00:00Z"
}
```

### 4. Check Stats

```bash
curl http://localhost:8080/stats
```

Response:
```json
{
  "total": 1,
  "by_status": {
    "RECEIVED": 1
  },
  "timestamp": "2025-01-30T12:00:05Z"
}
```

### 5. Wait ~5 seconds for worker to process

### 6. Check Stats Again

```bash
curl http://localhost:8080/stats
```

Response:
```json
{
  "total": 1,
  "by_status": {
    "CONFIRMED": 1
  },
  "timestamp": "2025-01-30T12:00:10Z"
}
```

### 7. Get Full Transaction

```bash
curl http://localhost:8080/transactions/0xtest123
```

Response (now with full data):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transaction_hash": "0xtest123",
  "chain_id": 1,
  "status": "CONFIRMED",
  "from_address": "0xabc...",
  "to_address": "0xdef...",
  "value": "1000000000000000000",
  "block_number": 12345678,
  "gas_used": 21000,
  "created_at": "2025-01-30T12:00:00Z",
  "updated_at": "2025-01-30T12:00:08Z"
}
```

## Configuration

Worker configuration in `internal/worker/processor.go`:

```go
type Worker struct {
    PollInterval  time.Duration  // How often to check for new transactions
    BatchSize     int            // Max transactions per batch
}
```

Default values:
- Poll interval: 5 seconds
- Batch size: 10 transactions

Adjust based on needs:
```go
worker := NewWorker(db)
worker.PollInterval = 1 * time.Second  // Poll every second
worker.BatchSize = 50                   // Process up to 50 at once
```

## Production Considerations

### What This Demonstrates

✅ **Async processing**: Decouples ingestion from processing
✅ **State machines**: Clean status transitions
✅ **Event logging**: Full audit trail in `ingestion_events`
✅ **Graceful shutdown**: Handles SIGTERM/SIGINT
✅ **Idempotency**: Can reprocess failed transactions
✅ **Monitoring**: Stats endpoint for observability

### What Would Be Added in Production

- **Real RPC calls**: Instead of simulation
- **Retry logic**: With exponential backoff
- **Dead letter queue**: For permanently failed transactions
- **Distributed locking**: If running multiple workers
- **Metrics**: Prometheus counters/histograms
- **Alerts**: On high error rates
- **Rate limiting**: To avoid hitting RPC limits
- **Circuit breaker**: When blockchain node is down

## Monitoring

### Check Worker Health

```bash
# Count pending transactions
psql $DATABASE_URL -c "SELECT COUNT(*) FROM transactions WHERE status = 'RECEIVED';"

# Check error rate
psql $DATABASE_URL -c "SELECT COUNT(*) FROM transactions WHERE status = 'ERROR';"

# Recent errors
psql $DATABASE_URL -c "
  SELECT transaction_hash, error_reason, updated_at 
  FROM transactions 
  WHERE status = 'ERROR' 
  ORDER BY updated_at DESC 
  LIMIT 10;
"
```

### Ingestion Events Log

```bash
# View recent status changes
psql $DATABASE_URL -c "
  SELECT 
    t.transaction_hash,
    e.previous_status,
    e.new_status,
    e.reason,
    e.created_at
  FROM ingestion_events e
  JOIN transactions t ON e.transaction_id = t.id
  ORDER BY e.created_at DESC
  LIMIT 20;
"
```

## Resume Description

**For your resume**:

> Designed and implemented an asynchronous transaction processing worker with state machine architecture (RECEIVED → FETCHING → CONFIRMED). Built polling-based background service with graceful shutdown, event logging, and monitoring endpoints. Demonstrated production patterns including idempotent processing, batch operations, and comprehensive audit trails.
