# Docker Quick Start Guide

This guide will get you up and running with TxnFlow in under 5 minutes.

## Prerequisites

- Docker installed ([Get Docker](https://docs.docker.com/get-docker/))
- Docker Compose installed (included with Docker Desktop)
- Terminal/Command Prompt
- `curl` (for testing) - optional but recommended

## 🚀 One-Command Startup

```bash
# Clone the repository
git clone https://github.com/Wuzu11517/txnflow.git
cd txnflow

# Start everything
docker-compose up -d
```

This single command will:
1. ✅ Pull PostgreSQL image
2. ✅ Build API and Worker images
3. ✅ Start PostgreSQL database
4. ✅ Run database migrations
5. ✅ Start API server on port 8080
6. ✅ Start background worker

**Wait ~10-15 seconds** for all services to initialize.

---

## ✅ Verify It's Running

### Check Service Status

```bash
docker-compose ps
```

You should see:
```
NAME                COMMAND              STATUS         PORTS
txnflow-api         "/app/api"           Up             0.0.0.0:8080->8080/tcp
txnflow-postgres    "docker-entrypoint…" Up             0.0.0.0:5432->5432/tcp
txnflow-worker      "/app/worker"        Up
```

### Check API Health

```bash
curl http://localhost:8080/stats
```

Expected response:
```json
{
  "total": 0,
  "by_status": {},
  "timestamp": "2025-01-30T12:00:00Z"
}
```

If you see this, everything is working! 🎉

---

## 🧪 Try It Out

### 1. Create a Transaction

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_hash": "0xabc123def456...",
    "chain_id": 1
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transaction_hash": "0xabc123def456...",
  "chain_id": 1,
  "status": "RECEIVED",
  "created_at": "2025-01-30T12:00:00Z"
}
```

### 2. Watch the Worker Process It

```bash
docker-compose logs -f worker
```

You should see:
```
📥 Processing transaction: 0xabc123def456... (chain: 1)
✅ Transaction 0xabc123def456... processed successfully
✅ Processed 1 transactions
```

Press `Ctrl+C` to exit logs.

### 3. Check Stats

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
  "timestamp": "2025-01-30T12:00:05Z"
}
```

Status changed from `RECEIVED` → `CONFIRMED`! ✅

### 4. Get the Transaction (Now With Full Data)

```bash
curl http://localhost:8080/transactions/0xabc123def456...
```

Response (now populated):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "transaction_hash": "0xabc123def456...",
  "chain_id": 1,
  "status": "CONFIRMED",
  "from_address": "0x1a2b3c4d5e6f...",
  "to_address": "0x7a8b9c0d1e2f...",
  "value": "1000000000000000000",
  "block_number": 12345678,
  "gas_used": 21000,
  "created_at": "2025-01-30T12:00:00Z",
  "updated_at": "2025-01-30T12:00:05Z"
}
```

---

## 🎲 Load Sample Data

Want to see it process multiple transactions?

```bash
# Using the Makefile (requires make)
make sample-data

# Or directly with the script
./scripts/load_sample_data.sh 20
```

This creates 20 transactions with random hashes. Watch them get processed:

```bash
make logs-worker
# or
docker-compose logs -f worker
```

Check stats again:
```bash
make stats
# or
curl http://localhost:8080/stats
```

---

## 📋 Common Commands

### Using Make (Recommended)

```bash
make up           # Start all services
make down         # Stop all services
make restart      # Restart all services
make logs         # View all logs
make logs-api     # View API server logs
make logs-worker  # View worker logs
make stats        # Check statistics
make sample-data  # Load sample transactions
make test         # Run basic API tests
make clean        # Stop and remove everything
```

### Using Docker Compose Directly

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f api
docker-compose logs -f worker
docker-compose logs -f postgres

# Rebuild after code changes
docker-compose build
docker-compose up -d

# Stop and remove volumes (clean slate)
docker-compose down -v
```

---

## 🔍 Exploring the API

### List All Transactions

```bash
curl http://localhost:8080/transactions?limit=10
```

### Filter by Status

```bash
curl "http://localhost:8080/transactions?status=CONFIRMED&limit=5"
```

### Filter by Chain

```bash
curl "http://localhost:8080/transactions?chain_id=1&limit=5"
```

### Pagination

```bash
# First page (first 10)
curl "http://localhost:8080/transactions?limit=10&offset=0"

# Second page (next 10)
curl "http://localhost:8080/transactions?limit=10&offset=10"
```

---

## 🐛 Troubleshooting

### Services Won't Start

**Check if ports are already in use:**
```bash
# Check if port 8080 is in use
lsof -i :8080

# Check if port 5432 is in use
lsof -i :5432
```

**Solution**: Stop other services using these ports, or edit `docker-compose.yml` to use different ports.

### API Returns Connection Errors

**Wait longer** - services need ~10-15 seconds to fully initialize.

**Check service health:**
```bash
docker-compose ps
```

All services should show `Up` status.

### Worker Not Processing Transactions

**Check worker logs:**
```bash
docker-compose logs worker
```

Look for errors. Common issues:
- Database connection problems
- Migration didn't run

**Restart worker:**
```bash
docker-compose restart worker
```

### Database Issues

**View database logs:**
```bash
docker-compose logs postgres
```

**Reset database (clean slate):**
```bash
docker-compose down -v
docker-compose up -d
```

**Note**: `-v` flag removes volumes, so all data will be lost.

---

## 🧹 Cleanup

### Stop Services (Keep Data)

```bash
docker-compose down
```

This stops containers but keeps the database volume.

### Complete Cleanup (Remove Everything)

```bash
docker-compose down -v
```

This removes:
- All containers
- All volumes (database data)
- Networks

Fresh slate next time you run `docker-compose up`.

### Remove Docker Images

```bash
docker-compose down -v --rmi all
```

Removes containers, volumes, AND images.

---

## 🎓 What's Happening Behind the Scenes

When you run `docker-compose up`:

1. **PostgreSQL container** starts
   - Creates `txnflow` database
   - Exposes port 5432

2. **Migration container** runs
   - Waits for PostgreSQL to be ready
   - Runs `001_init.sql` (creates tables)
   - Runs `002_indexes.sql` (creates indexes)
   - Exits after completion

3. **API container** starts
   - Waits for migrations to complete
   - Starts HTTP server on port 8080
   - Serves API endpoints

4. **Worker container** starts
   - Waits for API to be healthy
   - Starts polling for transactions
   - Processes them every 5 seconds

All containers run in an isolated network and can communicate with each other by name (e.g., `postgres`, `api`).

---

## 🚀 Next Steps

Now that you have it running:

1. **Explore the API** - try different filters and queries
2. **Watch the logs** - see the worker in action
3. **Check the code** - see how it's implemented
4. **Read the docs** - understand the architecture

Happy exploring! 🎉
