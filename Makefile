.PHONY: help build up down restart logs logs-api logs-worker logs-db stats sample-data clean test

# Default target
help:
	@echo "TxnFlow - Available Commands"
	@echo "============================"
	@echo ""
	@echo "  make up           - Start all services (PostgreSQL + API + Worker)"
	@echo "  make down         - Stop all services"
	@echo "  make restart      - Restart all services"
	@echo "  make build        - Rebuild Docker images"
	@echo ""
	@echo "  make logs         - View logs from all services"
	@echo "  make logs-api     - View API server logs"
	@echo "  make logs-worker  - View worker logs"
	@echo "  make logs-db      - View database logs"
	@echo ""
	@echo "  make stats        - Check transaction processing stats"
	@echo "  make sample-data  - Load 20 sample transactions"
	@echo ""
	@echo "  make clean        - Stop services and remove volumes"
	@echo "  make test         - Run basic API tests"
	@echo ""

# Build Docker images
build:
	@echo "🔨 Building Docker images..."
	docker-compose build

# Start all services
up:
	@echo "🚀 Starting TxnFlow services..."
	docker-compose up -d
	@echo ""
	@echo "✅ Services started!"
	@echo ""
	@echo "API Server: http://localhost:8080"
	@echo "Database:   localhost:5432"
	@echo ""
	@echo "Wait ~10 seconds for services to be ready, then try:"
	@echo "  make stats"
	@echo "  make sample-data"
	@echo ""

# Stop all services
down:
	@echo "⏹️  Stopping TxnFlow services..."
	docker-compose down

# Restart all services
restart: down up

# View all logs
logs:
	docker-compose logs -f

# View API logs
logs-api:
	docker-compose logs -f api

# View worker logs
logs-worker:
	docker-compose logs -f worker

# View database logs
logs-db:
	docker-compose logs -f postgres

# Check stats
stats:
	@echo "📊 Transaction Statistics:"
	@curl -s http://localhost:8080/stats | jq '.' || curl -s http://localhost:8080/stats

# Load sample data
sample-data:
	@./scripts/load_sample_data.sh 20

# Clean everything (including volumes)
clean:
	@echo "🧹 Cleaning up..."
	docker-compose down -v
	@echo "✅ Cleanup complete!"

# Basic API tests
test:
	@echo "🧪 Running basic API tests..."
	@echo ""
	@echo "1. Creating test transaction..."
	@curl -s -X POST http://localhost:8080/transactions \
		-H "Content-Type: application/json" \
		-d '{"transaction_hash":"0xtest123","chain_id":1}' | jq '.'
	@echo ""
	@echo "2. Checking stats..."
	@curl -s http://localhost:8080/stats | jq '.'
	@echo ""
	@echo "3. Listing transactions..."
	@curl -s "http://localhost:8080/transactions?limit=5" | jq '.'
	@echo ""
	@echo "✅ Tests complete!"
