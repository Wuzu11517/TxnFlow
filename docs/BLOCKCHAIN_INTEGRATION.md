# Real Blockchain Integration Setup

TxnFlow now integrates with **real Ethereum blockchain data** via Infura!

## 🔑 Getting Your Infura API Key (Free - 2 minutes)

### Step 1: Sign Up for Infura

1. Go to [https://infura.io](https://infura.io)
2. Click "Get Started for Free"
3. Sign up with your email
4. Verify your email address

### Step 2: Create an API Key

1. Log into your Infura dashboard
2. Click "Create New API Key"
3. Select "Web3 API (Formerly Ethereum)"
4. Give it a name like "TxnFlow Development"
5. Click "Create"

### Step 3: Copy Your API Key

1. Click on your new API key
2. Copy the **API Key** (looks like: `a1b2c3d4e5f6g7h8i9j0...`)
3. Save it - you'll need it in the next step!

**Free Tier Limits**:
- 100,000 requests/day
- More than enough for development and demos

---

## 🚀 Configure TxnFlow

### Option 1: Using .env File (Recommended)

1. Create a `.env` file in the project root:
```bash
cp .env.example .env
```

2. Edit `.env` and add your Infura API key:
```bash
INFURA_API_KEY=your_actual_api_key_here
```

3. Docker Compose will automatically load it:
```bash
docker-compose up -d
```

### Option 2: Set Environment Variable Directly

```bash
export INFURA_API_KEY=your_actual_api_key_here
docker-compose up -d
```

### Option 3: Inline with Docker Compose

```bash
INFURA_API_KEY=your_actual_api_key_here docker-compose up -d
```

---

## ✅ Verify It's Working

### Check Worker Logs

```bash
docker-compose logs worker
```

You should see:
```
✅ Connected to database: postgres://...
🔗 Supported chains: [1]
⚙️  Worker configuration:
   - Poll interval: 5s
   - Batch size: 10
   - Blockchain: Real Ethereum data via Infura
```

### Test with a Real Transaction

Use a real Ethereum transaction hash from [Etherscan](https://etherscan.io):

```bash
# Example: Real Ethereum transaction
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_hash": "0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060",
    "chain_id": 1
  }'
```

This is a **real transaction** from Ethereum mainnet! 

Wait ~5-10 seconds, then:

```bash
curl http://localhost:8080/transactions/0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060
```

You should see **real data** populated:
- Real from_address
- Real to_address  
- Real value (in wei)
- Real block_number
- Real gas_used

You can verify this matches [Etherscan](https://etherscan.io/tx/0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060)!

---

## 🔍 How to Find Test Transactions

### Method 1: Use Etherscan

1. Go to [https://etherscan.io](https://etherscan.io)
2. Look at "Latest Transactions"
3. Click on any transaction
4. Copy the "Transaction Hash"
5. Use it in your API call

### Method 2: Use Well-Known Transactions

Some famous Ethereum transactions you can test with:

```bash
# First Ethereum transaction ever
0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060

# Large ETH transfer
0x...
```

### Method 3: Create Your Own (If You Have ETH)

If you have an Ethereum wallet with some ETH:
1. Send a small transaction (even 0.001 ETH)
2. Get the transaction hash from your wallet
3. Test with your own real transaction!

---

## 🐛 Troubleshooting

### Error: "INFURA_API_KEY not set"

**Problem**: Worker can't start without API key

**Solution**: 
```bash
# Check if .env file exists
ls -la .env

# Check if key is set
grep INFURA_API_KEY .env

# Restart services
docker-compose restart worker
```

### Error: "transaction not found"

**Problem**: Transaction hash doesn't exist on Ethereum

**Possible reasons**:
- Typo in the hash
- Transaction is on a different chain (not Ethereum mainnet)
- Transaction is very recent and not yet confirmed

**Solution**: Use a confirmed transaction from Etherscan

### Error: "RPC call failed"

**Problem**: Can't connect to Infura

**Possible reasons**:
- Invalid API key
- Rate limit exceeded (100k requests/day)
- Infura service issue

**Solution**:
```bash
# Verify API key is correct
echo $INFURA_API_KEY

# Check Infura status
curl https://status.infura.io
```

### Worker Keeps Retrying

**Problem**: Transaction can't be fetched

**What to check**:
```bash
# View detailed error logs
docker-compose logs worker | grep ERROR

# Common issues:
# - Transaction is pending (not yet mined)
# - Transaction is on wrong chain
# - Invalid transaction hash format
```

---

## 📊 What Data Gets Fetched

For each transaction, TxnFlow fetches:

### From `eth_getTransactionByHash`:
- ✅ `hash` - Transaction hash
- ✅ `from` - Sender address
- ✅ `to` - Recipient address
- ✅ `value` - Amount transferred (in wei)
- ✅ `blockNumber` - Block it was included in
- ✅ `gas` - Gas limit
- ✅ `gasPrice` - Gas price

### From `eth_getTransactionReceipt`:
- ✅ `gasUsed` - Actual gas consumed
- ✅ `status` - Success (0x1) or Failed (0x0)
- ✅ `blockNumber` - Confirmation

### Stored in Database:
```sql
from_address:   VARCHAR  -- "0xabcd..."
to_address:     VARCHAR  -- "0x1234..."
value:          NUMERIC  -- "1000000000000000000" (1 ETH in wei)
block_number:   BIGINT   -- 12345678
gas_used:       BIGINT   -- 21000
```

---

## 🎯 Next Steps

Now that you have real blockchain integration:

1. ✅ **Test with multiple transactions** from Etherscan
2. ✅ **Verify data matches** on Etherscan
3. ✅ **Try different types** of transactions:
   - Simple ETH transfers
   - Smart contract interactions
   - Token transfers (will have `input` data)

---

## 🌐 Future: Multi-Chain Support

Your architecture is ready for multiple chains! To add Polygon or Arbitrum:

1. Uncomment in `internal/blockchain/registry.go`:
```go
// Polygon
registry.RegisterChain(&ChainConfig{
    ChainID: 137,
    Name:    "Polygon",
    RPCURL:  "https://polygon-rpc.com",  // Free, no API key needed!
    Type:    ChainTypeEVM,
})
```

2. Rebuild and restart:
```bash
docker-compose build
docker-compose up -d
```

3. Test with Polygon transaction:
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_hash": "0x...",
    "chain_id": 137
  }'
```

**Same code, different chain!** 🚀

---

## 📝 Summary

You now have:
- ✅ Real Ethereum blockchain integration
- ✅ Infura RPC connectivity
- ✅ Actual transaction data fetching
- ✅ Architecture ready for multi-chain expansion

This is no longer a simulation - **it's a real blockchain data pipeline!** 🎉
