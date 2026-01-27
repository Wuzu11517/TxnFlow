CREATE TYPE transaction_status AS ENUM (
  'RECEIVED',
  'FETCHING',
  'PENDING',
  'CONFIRMED',
  'FAILED',
  'DROPPED',
  'ERROR'
);

CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  transaction_hash TEXT NOT NULL,
  chain_id INTEGER NOT NULL,
  status transaction_status NOT NULL,
  from_address TEXT,
  to_address TEXT,
  value NUMERIC,
  block_number BIGINT,
  gas_used BIGINT,
  error_reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE (transaction_hash, chain_id)
);

CREATE TABLE ingestion_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  transaction_id UUID NOT NULL REFERENCES transactions(id),
  previous_status transaction_status,
  new_status transaction_status NOT NULL,
  reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);
