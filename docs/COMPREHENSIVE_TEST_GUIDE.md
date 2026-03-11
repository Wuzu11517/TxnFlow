# Comprehensive Test Suite Documentation

## Overview

The comprehensive test suite (`scripts/comprehensive-test.sh`) validates **all aspects** of TxnFlow including real blockchain integration, error handling, edge cases, performance, and data integrity.

---

## What It Tests (30 Tests Total)

### Category 1: Basic API Tests (5 tests)
- ✅ Stats endpoint returns valid JSON
- ✅ Create transaction with valid data
- ✅ Idempotent insert (duplicate handling)
- ✅ Get transaction by hash
- ✅ List transactions

### Category 2: Real Blockchain Integration (4 tests)
- ✅ Fetch and process real Ethereum transaction (first ETH tx)
- ✅ Validate blockchain data matches Etherscan
- ✅ Verify gas_used is populated from receipt
- ✅ Test with additional real transaction

### Category 3: Error Handling (5 tests)
- ✅ Handle nonexistent transaction hash
- ✅ Reject invalid JSON
- ✅ Reject missing required fields
- ✅ Return 404 for nonexistent transaction
- ✅ Handle unsupported chain ID

### Category 4: Edge Cases & Data Integrity (6 tests)
- ✅ Same hash on different chains (UNIQUE constraint)
- ✅ Filter by chain_id
- ✅ Filter by status
- ✅ Pagination (limit and offset)
- ✅ Large offset handling
- ✅ Excessively long transaction hash

### Category 5: Performance & Load Tests (4 tests)
- ✅ Batch create 50 transactions rapidly
- ✅ List large result set (100 items)
- ✅ Handle 10 concurrent requests
- ✅ Stats endpoint performance

### Category 6: Worker Processing (2 tests)
- ✅ Verify worker is actively processing
- ✅ Check status transitions (RECEIVED → FETCHING → CONFIRMED/ERROR)

### Category 7: Database Integrity (4 tests)
- ✅ Database connectivity
- ✅ Migrations applied
- ✅ Indexes exist
- ✅ No orphaned records

---

## How to Run

### Prerequisites

1. **Services must be running**:
```bash
docker-compose up -d
sleep 10  # Wait for services to initialize
```

2. **Infura API key must be set**:
```bash
# Check if set
docker-compose logs worker | grep "Real Ethereum data"

# If not, add to .env:
echo "INFURA_API_KEY=your_key_here" > .env
docker-compose restart worker
```

### Run All Tests

```bash
./scripts/comprehensive-test.sh
```

### Expected Runtime

- **Quick run**: ~2 minutes (if worker queue is empty)
- **Normal run**: ~3-5 minutes (with processing waits)
- **Slow run**: ~8 minutes (if network is slow)

---

## Understanding the Output

### Pre-Flight Checks

```
--- Pre-Flight Checks ---
ℹ️  INFO: Docker is running
ℹ️  INFO: docker-compose is available
ℹ️  INFO: All services are running (3 containers)
ℹ️  INFO: Worker configured with real Ethereum RPC
ℹ️  INFO: API is accessible
✅ All pre-flight checks passed
```

If any pre-flight check fails, the test exits immediately with an error message.

### Test Output Format

```
TEST 6: Fetch and process real Ethereum transaction (First ETH tx)
ℹ️  INFO: Using first ever Ethereum transaction: 0x5c504ed...
ℹ️  INFO: Verify on Etherscan: https://etherscan.io/tx/0x5c504ed...
✅ PASS: Real transaction created with RECEIVED status
ℹ️  INFO: Waiting for worker to fetch from Ethereum mainnet...
✅ PASS: Real transaction processed to CONFIRMED status
```

### Color Coding

- 🔵 **BLUE**: Headers and informational messages
- 🟢 **GREEN**: Passed tests
- 🔴 **RED**: Failed tests
- 🟡 **YELLOW**: Warnings (not failures, but noteworthy)
- 🔴 **MAGENTA**: Critical errors

### Pass/Fail/Warn States

**PASS** ✅: Test succeeded
```
✅ PASS: Transaction created with RECEIVED status
```

**FAIL** ❌: Test failed (with details)
```
❌ FAIL: Invalid transaction not marked as ERROR
   Error: Expected status ERROR, got RECEIVED
```

**WARN** ⚠️: Not a failure, but something to check
```
⚠️  WARN: Transaction still in FETCHING state (may need more time)
```

---

## Test Details

### Test 6-8: Real Blockchain Validation

These are the **most important tests** - they prove you're using real data!

**Test 6**: Creates transaction with hash `0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060`
- This is the **first ever Ethereum transaction**
- Waits 15 seconds for worker to fetch from Infura
- Checks status becomes CONFIRMED

**Test 7**: Validates the fetched data matches Etherscan
```
Expected Values (from Etherscan):
- from_address: 0xa1e4380a3b1f749673e270229993ee55f35663b4
- to_address:   0x5df9b87991262f6ba471f09758cde1c0fc1de734
- value:        31337 wei
- block_number: 46147
```

If all 4 values match exactly, you're fetching **real blockchain data**! 🎉

**Test 8**: Checks that `gas_used` is populated from the receipt
- Should be 21000 (standard ETH transfer)

### Test 10: Error Handling (Invalid Hash)

Creates a fake transaction hash and verifies:
1. It's accepted initially (status: RECEIVED)
2. Worker attempts to fetch it
3. Fails with RPC error
4. Marks transaction as ERROR
5. Error reason is stored

### Test 21-24: Performance Tests

**Test 21**: Creates 50 transactions as fast as possible
- Measures total time
- Calculates requests/second
- Example: "Created 50 transactions in 3s (~16 req/sec)"

**Test 22**: Lists 100 transactions and measures latency
- Good: <100ms
- Acceptable: <500ms
- Slow: >500ms (may need index tuning)

**Test 23**: Fires 10 concurrent POST requests
- Verifies all succeed
- Tests connection pool handling

**Test 24**: Measures stats endpoint performance
- Should be very fast (<50ms) due to index on status column

### Test 27-30: Database Integrity

**Test 27**: Verifies PostgreSQL is accessible

**Test 28**: Checks migrations created required tables:
- `transactions`
- `ingestion_events`

**Test 29**: Counts indexes
- Expected: 6+ indexes on transactions table
- If less, indexing may be incomplete

**Test 30**: Checks for orphaned ingestion_events
- Events should always reference valid transaction_id
- If orphans exist, foreign key constraint may be missing

---

## Final Results

### All Tests Pass ✅

```
============================================================================
🎉 ALL TESTS PASSED! 🎉
============================================================================

✅ Real blockchain integration working
✅ Error handling robust
✅ Database integrity maintained
✅ Performance acceptable

Your TxnFlow system is production-ready! 🚀
```

Exit code: `0`

### Some Tests Fail ❌

```
============================================================================
⚠️  SOME TESTS FAILED ⚠️
============================================================================

Review failed tests above and check:
  - docker-compose logs worker
  - docker-compose logs api
  - docker-compose logs postgres
```

Exit code: `1`

---

## Troubleshooting Failed Tests

### Test 6-7 Fail: Blockchain Integration

**Symptom**:
```
❌ FAIL: Real transaction processed to CONFIRMED status
   Transaction still in RECEIVED state
```

**Possible Causes**:
1. INFURA_API_KEY not set correctly
2. Infura rate limit exceeded
3. Network connectivity issues
4. Worker not running

**Fix**:
```bash
# Check worker logs
docker-compose logs worker | grep -i error

# Verify Infura key
grep INFURA_API_KEY .env

# Test Infura directly
curl -X POST https://mainnet.infura.io/v3/YOUR_KEY \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Should return: {"jsonrpc":"2.0","id":1,"result":"0x..."}
```

### Test 7 Fails: Data Validation

**Symptom**:
```
❌ FAIL: from_address mismatch
   Expected: 0xa1e4380a..., Got: null
```

**Possible Cause**: Worker processed but didn't populate fields

**Fix**:
```bash
# Check worker logs for parsing errors
docker-compose logs worker | grep "⚠️"

# Look for:
# ⚠️  Failed to parse value: ...
# ⚠️  Failed to parse block number: ...

# This indicates hex parsing issue
# Check internal/blockchain/utils.go is present
ls internal/blockchain/utils.go
```

### Test 10 Fails: Error Handling

**Symptom**:
```
❌ FAIL: Invalid transaction not marked as ERROR
```

**Possible Cause**: Worker not processing, or timeout too short

**Fix**:
```bash
# Increase wait time in test
# Edit scripts/comprehensive-test.sh
# Change: wait_for_processing 10
# To:     wait_for_processing 20
```

### Test 21 Fails: Batch Creation

**Symptom**:
```
❌ FAIL: Only created 35/50 transactions
```

**Possible Causes**:
1. API rate limiting
2. Database connection pool exhausted
3. Network timeout

**Fix**:
```bash
# Check API logs
docker-compose logs api | tail -50

# Look for:
# - Connection pool errors
# - Timeout errors
# - 500 errors

# Increase connection pool (in config or docker-compose)
# Or reduce batch size in test
```

### Test 27-28 Fail: Database Issues

**Symptom**:
```
❌ FAIL: Database connection failed
```

**Fix**:
```bash
# Check postgres is running
docker-compose ps postgres

# Check postgres logs
docker-compose logs postgres

# Try connecting manually
docker exec -it txnflow-postgres psql -U txnflow -d txnflow

# If connection fails, restart postgres
docker-compose restart postgres
```

---

## Customizing the Test Suite

### Skip Slow Tests

Edit the script and comment out performance tests:

```bash
# Comment out TEST 21-24
# next_test
# PERFORMANCE_TESTS=$((PERFORMANCE_TESTS + 1))
# print_test "Create 50 transactions rapidly"
# ... (comment all of test 21)
```

### Add Your Own Tests

Add after TEST 30:

```bash
# TEST 31: Your custom test
next_test
print_test "Your test description"

# Your test logic here
RESULT=$(curl -s "$API_URL/your-endpoint")

if echo "$RESULT" | grep -q "expected"; then
    pass "Your test passed"
else
    fail "Your test failed" "$RESULT"
fi
```

### Change Wait Times

If your system is slow, increase processing waits:

```bash
# Find all instances of:
wait_for_processing 10

# Change to:
wait_for_processing 20
```

---

## What Makes This Comprehensive

Unlike basic tests, this suite checks:

### ✅ Real Blockchain Integration
- Not just "does it work"
- **Validates data matches Etherscan exactly**
- Tests receipt fetching (gas_used)
- Multiple real transactions

### ✅ Edge Cases
- Same hash different chains
- Invalid/nonexistent hashes
- Unsupported chain IDs
- Extremely long inputs
- Large offsets

### ✅ Performance Under Load
- Batch inserts
- Concurrent requests
- Large queries
- Response times measured

### ✅ Database Integrity
- Tables exist
- Indexes exist
- No orphaned records
- Foreign key constraints

### ✅ Worker Behavior
- Actually processing
- State transitions correct
- Error handling works
- Logs show activity

### ✅ API Robustness
- Invalid JSON rejected
- Missing fields rejected
- 404 for nonexistent
- Pagination works

---

## CI/CD Integration

### GitHub Actions

```yaml
name: TxnFlow Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v2
      
      - name: Create .env
        run: echo "INFURA_API_KEY=${{ secrets.INFURA_API_KEY }}" > .env
      
      - name: Start services
        run: docker-compose up -d
      
      - name: Wait for services
        run: sleep 15
      
      - name: Run comprehensive tests
        run: ./scripts/comprehensive-test.sh
      
      - name: Cleanup
        if: always()
        run: docker-compose down -v
```

### Jenkins

```groovy
pipeline {
    agent any
    
    environment {
        INFURA_API_KEY = credentials('infura-api-key')
    }
    
    stages {
        stage('Setup') {
            steps {
                sh 'echo "INFURA_API_KEY=$INFURA_API_KEY" > .env'
                sh 'docker-compose up -d'
                sh 'sleep 15'
            }
        }
        
        stage('Test') {
            steps {
                sh './scripts/comprehensive-test.sh'
            }
        }
        
        stage('Cleanup') {
            steps {
                sh 'docker-compose down -v'
            }
        }
    }
}
```

---

## Quick Reference

```bash
# Run all tests
./scripts/comprehensive-test.sh

# Run with verbose output
bash -x ./scripts/comprehensive-test.sh

# Save results to file
./scripts/comprehensive-test.sh | tee test-results.log

# Run only pre-flight checks (debug mode)
./scripts/comprehensive-test.sh | head -50

# Check exit code
./scripts/comprehensive-test.sh
echo $?  # 0 = all passed, 1 = failures
```

---

## Expected Output Summary

**When Everything Works**:

```
Total Tests:  30
Passed:       30
Failed:       0
Warnings:     0-2 (warnings are ok)

Pass Rate:    100.0%

🎉 ALL TESTS PASSED! 🎉
```

**With Minor Issues**:

```
Total Tests:  30
Passed:       28
Failed:       0
Warnings:     2

Pass Rate:    93.3%

✅ Real blockchain integration working
⚠️  Some performance tests slower than ideal
```

**With Failures**:

```
Total Tests:  30
Passed:       25
Failed:       5
Warnings:     3

Pass Rate:    83.3%

⚠️  SOME TESTS FAILED ⚠️
Review failed tests above
```

---

This test suite is **production-grade** and covers scenarios you might not think of during development. Run it before any demo or deployment! 🚀
