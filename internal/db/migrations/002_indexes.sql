-- Index 1: Transaction hash lookup (supports GET by hash)
-- Note: UNIQUE (transaction_hash, chain_id) already exists from 001_init.sql
-- This standalone index optimizes queries filtering by hash alone
CREATE INDEX IF NOT EXISTS idx_transactions_hash 
ON transactions (transaction_hash);

-- Index 2: From address + block number (supports sender-based queries)
-- Composite index enables efficient filtering by:
--   - from_address alone
--   - from_address + block_number together
CREATE INDEX IF NOT EXISTS idx_transactions_from_block 
ON transactions (from_address, block_number) 
WHERE from_address IS NOT NULL;

-- Index 3: To address + block number (supports recipient-based queries)
-- Composite index enables efficient filtering by:
--   - to_address alone
--   - to_address + block_number together
CREATE INDEX IF NOT EXISTS idx_transactions_to_block 
ON transactions (to_address, block_number) 
WHERE to_address IS NOT NULL;

-- Index 4: Block number (supports block-based queries)
-- Used when filtering by block number without address filters
CREATE INDEX IF NOT EXISTS idx_transactions_block_number 
ON transactions (block_number) 
WHERE block_number IS NOT NULL;

-- Index 5: Chain ID + timestamp (supports chain-specific queries with time ordering)
CREATE INDEX IF NOT EXISTS idx_transactions_chain_created 
ON transactions (chain_id, created_at DESC);

-- Index 6: Status + timestamp (supports filtering by transaction status)
CREATE INDEX IF NOT EXISTS idx_transactions_status_created 
ON transactions (status, created_at DESC);