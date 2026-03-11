# TxnFlow

> Blockchain transaction processing pipeline with async worker pattern and real Ethereum integration.

---

## What It Does

TxnFlow demonstrates the transaction indexing architecture used by crypto wallets (MetaMask), exchanges (Coinbase), and DeFi apps. Submit a transaction hash → worker fetches blockchain data → stores in PostgreSQL → query instantly.

**Why not just use Etherscan API?** Rate limits (5 calls/sec free), cost ($499/month paid), no custom logic, privacy concerns. Every serious crypto app builds their own indexing infrastructure.

---

## Features

✅ Async worker pattern (non-blocking)  
✅ Real Ethereum data via Infura  
✅ Strategic database indexing (<5ms queries)  
✅ Rate limiting (100 req/hour per IP)  
✅ Production deployment (Fly.io + Vercel)  

---

## Tech Stack

**Backend**: Go, PostgreSQL, Docker, Chi Router  
**Frontend**: Vanilla JS, Tailwind CSS  
**Infrastructure**: Fly.io, Vercel, Infura  

---

## Quick Start

```bash
# Clone and start
git clone https://github.com/yourusername/txnflow.git
cd txnflow
docker-compose up -d

# Test
curl http://localhost:8080/stats

# Submit real Ethereum transaction
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"transaction_hash":"0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060","chain_id":1}'

# Watch it process (RECEIVED → FETCHING → CONFIRMED)
curl http://localhost:8080/transactions/0x5c504...
```

---

## Architecture

```
Frontend (Vercel) 
    ↓
API Server (Fly.io) - Go + Chi Router
    ↓
PostgreSQL (Fly.io) - 6 strategic indexes
    ↑
Background Worker (Fly.io) - Fetches from Ethereum
    ↓
Infura/Alchemy - Ethereum RPC
```

**Key Pattern**: API stores hash immediately (fast) → Worker fetches blockchain data in background → Updates database with full details.

---

## API Endpoints

```bash
POST /transactions      # Submit transaction hash
GET  /transactions      # List with filters (?status=CONFIRMED&limit=10)
GET  /transactions/:hash # Get transaction details
GET  /stats            # System statistics
```

---

## Performance

| Metric | Value |
|--------|-------|
| API Throughput | 1,200+ req/sec |
| Query Latency | <5ms (indexed) |
| Worker Rate | 6-10 tx/min |

---

## Future Plans

- Multi-chain support (Polygon, Arbitrum)
- Smart contract event decoding
- Portfolio tracking
- WebSocket real-time updates

---

## License

MIT

---