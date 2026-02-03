#!/bin/bash
set -e

# ============================================================================
# Database Migration Runner
# ============================================================================
# Usage: ./scripts/migrate.sh [DATABASE_URL]
#
# Example:
#   ./scripts/migrate.sh postgres://localhost/txnflow?sslmode=disable
#
# If DATABASE_URL is not provided, will use environment variable or default
# ============================================================================

DATABASE_URL="${1:-${DATABASE_URL:-postgres://localhost/txnflow?sslmode=disable}}"

echo "============================================================================"
echo "TxnFlow Database Migration Runner"
echo "============================================================================"
echo "Target: $DATABASE_URL"
echo ""

MIGRATIONS_DIR="internal/db/migrations"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "❌ Error: migrations directory not found at $MIGRATIONS_DIR"
    exit 1
fi

# Get sorted list of .sql files
MIGRATION_FILES=$(ls -1 $MIGRATIONS_DIR/*.sql 2>/dev/null | sort)

if [ -z "$MIGRATION_FILES" ]; then
    echo "⚠️  No migration files found in $MIGRATIONS_DIR"
    exit 0
fi

echo "Found migrations:"
for file in $MIGRATION_FILES; do
    echo "  - $(basename $file)"
done
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "❌ Error: psql command not found"
    echo "   Please install PostgreSQL client tools"
    exit 1
fi

# Apply each migration
for file in $MIGRATION_FILES; do
    filename=$(basename $file)
    echo "📦 Applying: $filename"
    
    if psql "$DATABASE_URL" -f "$file" -v ON_ERROR_STOP=1 --quiet; then
        echo "   ✅ Success"
    else
        echo "   ❌ Failed to apply $filename"
        exit 1
    fi
done

echo ""
echo "============================================================================"
echo "✅ All migrations applied successfully!"
echo "============================================================================"
echo ""
echo "Verifying indexes..."
echo ""

# Verify indexes were created
psql "$DATABASE_URL" -c "
SELECT 
    schemaname,
    tablename, 
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename IN ('transactions', 'ingestion_events')
ORDER BY tablename, indexname;
" || echo "⚠️  Could not verify indexes"

echo ""
echo "🎉 Migration complete!"
