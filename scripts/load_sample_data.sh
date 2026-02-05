#!/bin/bash
set -e

# ============================================================================
# Sample Data Loader
# ============================================================================
# Creates sample transactions to demonstrate the system
# Usage: ./scripts/load_sample_data.sh [count]
# ============================================================================

API_URL="${API_URL:-http://localhost:8080}"
COUNT="${1:-20}"

echo "============================================================================"
echo "TxnFlow Sample Data Loader"
echo "============================================================================"
echo "API URL: $API_URL"
echo "Creating $COUNT sample transactions..."
echo ""

# Function to generate random hex
random_hex() {
    openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p -c 64
}

# Create transactions
for i in $(seq 1 $COUNT); do
    HASH="0x$(random_hex)"
    CHAIN_ID=$((1 + RANDOM % 3))  # Random chain ID 1-3
    
    echo "[$i/$COUNT] Creating transaction: $HASH (chain: $CHAIN_ID)"
    
    curl -s -X POST "$API_URL/transactions" \
        -H "Content-Type: application/json" \
        -d "{
            \"transaction_hash\": \"$HASH\",
            \"chain_id\": $CHAIN_ID,
            \"source_service\": \"sample-data-loader\"
        }" > /dev/null
    
    # Small delay to avoid overwhelming the API
    sleep 0.1
done

echo ""
echo "============================================================================"
echo "✅ Created $COUNT transactions"
echo "============================================================================"
echo ""
echo "Wait a few seconds for the worker to process them, then check:"
echo "  curl $API_URL/stats"
echo ""
echo "Or view transactions:"
echo "  curl $API_URL/transactions?limit=10"
echo ""
