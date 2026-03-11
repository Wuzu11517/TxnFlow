# Database Indexing Strategy

## Overview

This document explains the indexing strategy for the TxnFlow transaction ingestion service. Proper indexing is **critical** for maintaining O(log n) query performance as the transaction table grows.

## Current Indexes

### From 001_init.sql

| Index Type | Columns | Purpose |
|------------|---------|---------|
| PRIMARY KEY | `id` | Unique row identifier |
| UNIQUE | `(transaction_hash, chain_id)` | Prevents duplicate transactions per chain |

### From 002_indexes.sql

| Index Name | Columns | Query Pattern | Partial? |
|------------|---------|---------------|----------|
| `idx_transactions_hash` | `transaction_hash` | GET /transactions/{hash} | No |
| `idx_transactions_from_block` | `from_address, block_number` | Filter by sender ± block | Yes (WHERE from_address IS NOT NULL) |
| `idx_transactions_to_block` | `to_address, block_number` | Filter by recipient ± block | Yes (WHERE to_address IS NOT NULL) |
| `idx_transactions_block_number` | `block_number` | Filter by block only | Yes (WHERE block_number IS NOT NULL) |
| `idx_transactions_chain_created` | `chain_id, created_at DESC` | Chain-specific + time ordering | No |
| `idx_transactions_status_created` | `status, created_at DESC` | Status filtering + time ordering | No |

## Index Design Rationale

### 1. Transaction Hash Lookup (`idx_transactions_hash`)

**Query**: `GET /transactions/{hash}`

```sql
SELECT * FROM transactions WHERE transaction_hash = $1
```

**Why this index**:
- Most common read operation
- O(log n) lookup instead of O(n) table scan
- Although `UNIQUE (transaction_hash, chain_id)` exists, a standalone hash index is more optimal when chain_id is unknown

**Performance impact**:
- Without index: ~1000ms for 10M rows
- With index: ~1ms for 10M rows

### 2. Address + Block Composite Indexes

**Queries**:
```sql
-- Sender-based
SELECT * FROM transactions WHERE from_address = $1
SELECT * FROM transactions WHERE from_address = $1 AND block_number > $2

-- Recipient-based  
SELECT * FROM transactions WHERE to_address = $1
SELECT * FROM transactions WHERE to_address = $1 AND block_number < $2
```

**Why composite indexes**:
- PostgreSQL can use leftmost prefix → `from_address` alone uses the index
- Block number range scans are efficient when address is specified first
- Order matters: `(from_address, block_number)` enables both queries above

**Why partial indexes** (`WHERE ... IS NOT NULL`):
- Transactions in RECEIVED/FETCHING state may not have addresses yet
- Partial indexes are smaller and faster
- Only index rows where the column is actually queryable

### 3. Block Number Index (`idx_transactions_block_number`)

**Query**:
```sql
SELECT * FROM transactions WHERE block_number = $1
SELECT * FROM transactions WHERE block_number BETWEEN $1 AND $2
```

**Why separate from composite**:
- Queries filtering by block alone (no address) need dedicated index
- Range scans on block_number are common (e.g., "all txns in block 12345")

### 4. Chain + Timestamp Index (`idx_transactions_chain_created`)

**Query**:
```sql
SELECT * FROM transactions 
WHERE chain_id = $1 
ORDER BY created_at DESC 
LIMIT 100
```

**Why `created_at DESC`**:
- Most common query pattern: "show me recent transactions"
- DESC index enables efficient reverse ordering without sort step
- PostgreSQL can scan index backward for ORDER BY ... DESC

### 5. Status + Timestamp Index (`idx_transactions_status_created`)

**Query**:
```sql
SELECT * FROM transactions 
WHERE status = 'PENDING' 
ORDER BY created_at DESC
```

**Use cases**:
- Admin dashboards: "show failed transactions"
- Monitoring: "count pending transactions"
- Processing queues: "get oldest RECEIVED transactions"

## Index Maintenance

### Size Monitoring

```sql
-- Check index sizes
SELECT 
    indexname,
    pg_size_pretty(pg_relation_size(indexname::regclass)) AS size
FROM pg_indexes 
WHERE tablename = 'transactions';
```

### Usage Statistics

```sql
-- Check if indexes are actually being used
SELECT 
    indexrelname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public' AND relname = 'transactions';
```

### Bloat Management

Over time, indexes can become bloated. Rebuild if needed:

```sql
REINDEX INDEX idx_transactions_hash;
-- Or rebuild all indexes on table:
REINDEX TABLE transactions;
```

## Query Optimizer Examples

Use `EXPLAIN ANALYZE` to verify index usage:

```sql
-- Should use idx_transactions_hash
EXPLAIN ANALYZE
SELECT * FROM transactions WHERE transaction_hash = '0xabc123';

-- Should use idx_transactions_from_block  
EXPLAIN ANALYZE
SELECT * FROM transactions WHERE from_address = '0xdef456';

-- Should use idx_transactions_chain_created
EXPLAIN ANALYZE
SELECT * FROM transactions 
WHERE chain_id = 1 
ORDER BY created_at DESC 
LIMIT 100;
```

Look for:
- ✅ `Index Scan using idx_transactions_XXX` (good!)
- ❌ `Seq Scan on transactions` (bad - full table scan)

## Performance Targets

| Operation | Target Latency | Index Required |
|-----------|---------------|----------------|
| GET by hash | < 10ms | idx_transactions_hash |
| Filter by address | < 50ms | idx_transactions_from/to_block |
| Filter by block | < 50ms | idx_transactions_block_number |
| List recent (paginated) | < 100ms | idx_transactions_chain_created |

## Trade-offs

**Benefits**:
- ✅ Sub-millisecond lookups
- ✅ Scalable to billions of rows
- ✅ Predictable performance

**Costs**:
- ❌ ~30-50% additional disk space (6 indexes × table size)
- ❌ Slower inserts (~10-20% overhead per index)
- ❌ Maintenance (VACUUM, REINDEX)

**Verdict**: For a read-heavy analytics system, this is the right trade-off.

## Next Steps

1. Apply migrations:
   ```bash
   psql $DATABASE_URL < internal/db/migrations/001_init.sql
   psql $DATABASE_URL < internal/db/migrations/002_indexes.sql
   ```

2. Load test data and verify with `EXPLAIN ANALYZE`

3. Monitor `pg_stat_user_indexes` in production

4. Consider adding more indexes as query patterns emerge
