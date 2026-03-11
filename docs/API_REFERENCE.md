# API Reference

## Base URL

```
http://localhost:8080
```

---

## Endpoints

### 1. Create Transaction

**Endpoint**: `POST /transactions`

**Description**: Register a new transaction for ingestion and tracking.

**Request Body**:
```json
{
  "transaction_hash": "0xabc123...",
  "chain_id": 1,
  "source_service": "indexer-v1"
}
```

**Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transaction_hash": "0xabc123...",
  "chain_id": 1,
  "status": "RECEIVED",
  "created_at": "2025-01-29T18:45:00Z"
}
```

**Behavior**:
- Idempotent: Duplicate `(transaction_hash, chain_id)` pairs return existing record
- Initial status: `RECEIVED`
- Creates an ingestion event log entry

**Error Responses**:
- `400 Bad Request`: Missing required fields or invalid JSON
- `500 Internal Server Error`: Database error

---

### 2. Get Transaction by Hash

**Endpoint**: `GET /transactions/{hash}`

**Description**: Retrieve a single transaction by its hash.

**Parameters**:
- `hash` (path, required): Transaction hash (e.g., `0xabc123...`)

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transaction_hash": "0xabc123...",
  "chain_id": 1,
  "status": "CONFIRMED",
  "from_address": "0xdef456...",
  "to_address": "0x789abc...",
  "value": "1000000000000000000",
  "block_number": 12345678,
  "gas_used": 21000,
  "error_reason": null,
  "created_at": "2025-01-29T18:45:00Z",
  "updated_at": "2025-01-29T18:46:30Z"
}
```

**Error Responses**:
- `400 Bad Request`: Missing hash parameter
- `404 Not Found`: Transaction not found
- `500 Internal Server Error`: Database error

**Performance**:
- Uses `idx_transactions_hash` index
- Target latency: < 10ms

---

### 3. List Transactions (with Filtering)

**Endpoint**: `GET /transactions`

**Description**: Query transactions with optional filters and pagination.

**Query Parameters**:

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `from_address` | string | Filter by sender address | `?from_address=0xdef456...` |
| `to_address` | string | Filter by recipient address | `?to_address=0x789abc...` |
| `chain_id` | integer | Filter by blockchain | `?chain_id=1` |
| `status` | string | Filter by transaction status | `?status=CONFIRMED` |
| `block_number_min` | integer | Minimum block number (inclusive) | `?block_number_min=12345000` |
| `block_number_max` | integer | Maximum block number (inclusive) | `?block_number_max=12346000` |
| `limit` | integer | Results per page (1-1000, default: 100) | `?limit=50` |
| `offset` | integer | Number of results to skip (default: 0) | `?offset=100` |

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "transaction_hash": "0xabc123...",
      "chain_id": 1,
      "status": "CONFIRMED",
      "from_address": "0xdef456...",
      "to_address": "0x789abc...",
      "value": "1000000000000000000",
      "block_number": 12345678,
      "gas_used": 21000,
      "error_reason": null,
      "created_at": "2025-01-29T18:45:00Z",
      "updated_at": "2025-01-29T18:46:30Z"
    }
  ],
  "limit": 100,
  "offset": 0,
  "count": 1
}
```

**Sorting**:
- Results ordered by `created_at DESC` (newest first)

**Example Queries**:

```bash
# Get all transactions from a specific address
GET /transactions?from_address=0xdef456...

# Get transactions in a block range
GET /transactions?block_number_min=12345000&block_number_max=12346000

# Get failed transactions on Ethereum mainnet
GET /transactions?chain_id=1&status=FAILED

# Paginated results
GET /transactions?limit=50&offset=100

# Complex filter: transactions from address A to address B in a block range
GET /transactions?from_address=0xaaa...&to_address=0xbbb...&block_number_min=12345000
```

**Performance**:
- Uses appropriate indexes based on filters:
  - `from_address` → `idx_transactions_from_block`
  - `to_address` → `idx_transactions_to_block`
  - `block_number` → `idx_transactions_block_number`
  - `chain_id` → `idx_transactions_chain_created`
  - `status` → `idx_transactions_status_created`
- Target latency: < 100ms for typical queries

**Error Responses**:
- `500 Internal Server Error`: Database error

---

## Transaction Status Enum

Valid values for the `status` field:

| Status | Description |
|--------|-------------|
| `RECEIVED` | Transaction hash received, not yet fetched |
| `FETCHING` | Fetching transaction data from blockchain |
| `PENDING` | Transaction is pending in mempool |
| `CONFIRMED` | Transaction confirmed on blockchain |
| `FAILED` | Transaction failed/reverted |
| `DROPPED` | Transaction dropped from mempool |
| `ERROR` | Error occurred during processing |

---

## Pagination

All list endpoints support offset-based pagination:

```bash
# Page 1 (results 0-99)
GET /transactions?limit=100&offset=0

# Page 2 (results 100-199)
GET /transactions?limit=100&offset=100

# Page 3 (results 200-299)
GET /transactions?limit=100&offset=200
```

**Limits**:
- Maximum `limit`: 1000
- Default `limit`: 100
- Minimum `offset`: 0

---

## Common Use Cases

### 1. Track transaction lifecycle
```bash
# Create
POST /transactions
{"transaction_hash": "0xabc...", "chain_id": 1}

# Poll status
GET /transactions/0xabc...
```

### 2. Address activity monitoring
```bash
# All outgoing transactions from address
GET /transactions?from_address=0xdef456...

# All incoming transactions to address  
GET /transactions?to_address=0xdef456...
```

### 3. Block explorer queries
```bash
# All transactions in block
GET /transactions?block_number_min=12345678&block_number_max=12345678&chain_id=1

# Recent activity on chain
GET /transactions?chain_id=1&limit=100
```

### 4. Error monitoring
```bash
# Failed transactions
GET /transactions?status=FAILED&limit=50

# Errors in recent blocks
GET /transactions?status=ERROR&block_number_min=12345000
```

---

## Performance Tips

1. **Always specify chain_id when possible** - enables efficient index usage
2. **Use block_number ranges for historical queries** - more efficient than timestamp filters
3. **Limit result sets** - default 100, max 1000
4. **Combine filters** - `from_address + block_number` uses composite index efficiently
5. **Consider caching** - frequently accessed transactions can be cached

---

## Database Indexes

The following indexes support fast queries:

| Query Pattern | Index Used |
|---------------|------------|
| `transaction_hash = X` | `idx_transactions_hash` |
| `from_address = X` | `idx_transactions_from_block` |
| `to_address = X` | `idx_transactions_to_block` |
| `block_number BETWEEN X AND Y` | `idx_transactions_block_number` |
| `chain_id = X ORDER BY created_at` | `idx_transactions_chain_created` |
| `status = X ORDER BY created_at` | `idx_transactions_status_created` |

See [INDEXING_STRATEGY.md](./INDEXING_STRATEGY.md) for details.
