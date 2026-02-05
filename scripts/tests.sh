#!/bin/bash

echo "=== TxnFlow Test Suite ==="
echo ""

echo "1. Checking API health..."
curl -s http://localhost:8080/stats | grep -q "total" && echo "✅ API is running" || echo "❌ API not responding"

echo ""
echo "2. Creating test transaction..."
RESPONSE=$(curl -s -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"transaction_hash":"0xtest123","chain_id":1}')
echo "$RESPONSE" | grep -q "RECEIVED" && echo "✅ Transaction created" || echo "❌ Failed to create"

echo ""
echo "3. Waiting 10 seconds for worker to process..."
sleep 10

echo ""
echo "4. Checking if transaction was processed..."
curl -s http://localhost:8080/stats | grep -q "CONFIRMED" && echo "✅ Worker processed transaction" || echo "⚠️  Still processing"

echo ""
echo "5. Getting transaction details..."
curl -s http://localhost:8080/transactions/0xtest123 | grep -q "from_address" && echo "✅ Transaction fully populated" || echo "❌ Transaction not populated"

echo ""
echo "=== Test Complete ==="