package http

import (
	"context",
	"time",
	"encoding/json",
	"net/http",

	"github.com/jackc/pgx/v5/pgxpool",
)

type Handlers struct {
	DB *pgxpool.Pool
}

func NewHandlers(db *pgxpool.Pool) *Handlers {
	return &Handlers{DB: db}
}

type createTransactionRequest struct {
	TransactionHash string `json:"transaction_hash"`
	ChainID         int    `json:"chain_id"`
	SourceService   string `json:"source_service"`
}

func (h *Handlers) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	// ---- Decode request body ----
	var req createTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// ---- Basic validation ----
	if req.TransactionHash == "" || req.ChainID == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// ---- Request-scoped context with timeout ----
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// ---- Attempt idempotent insert ----
	insertQuery := `
		INSERT INTO transactions (transaction_hash, chain_id, status, created_at, updated_at)
		VALUES ($1, $2, 'RECEIVED', now(), now())
		ON CONFLICT (transaction_hash, chain_id)
		DO NOTHING
		RETURNING id, status, created_at, updated_at
	`

	var (
		id        string
		status    string
		createdAt time.Time
		updatedAt time.Time
	)

	err := h.db.QueryRow(
		ctx,
		insertQuery,
		req.TransactionHash,
		req.ChainID,
	).Scan(&id, &status, &createdAt, &updatedAt)

	// ---- If insert did nothing, fetch existing row ----
	if err != nil {
		selectQuery := `
			SELECT id, status, created_at, updated_at
			FROM transactions
			WHERE transaction_hash = $1 AND chain_id = $2
		`

		row := h.db.QueryRow(
			ctx,
			selectQuery,
			req.TransactionHash,
			req.ChainID,
		)

		if err := row.Scan(&id, &status, &createdAt, &updatedAt); err != nil {
			http.Error(w, "failed to fetch transaction", http.StatusInternalServerError)
			return
		}
	}

	// ---- Record ingestion event (best-effort) ----
	_, _ = h.db.Exec(
		ctx,
		`INSERT INTO ingestion_events (transaction_id, new_status, reason)
		 VALUES ($1, $2, $3)`,
		id,
		status,
		"transaction registered",
	)

	// ---- Build response ----
	resp := map[string]interface{}{
		"id":               id,
		"transaction_hash": req.TransactionHash,
		"chain_id":         req.ChainID,
		"status":           status,
		"created_at":       createdAt,
	}

	// ---- Write response ----
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

type Transaction struct {
	ID              string    `json:"id"`
	TransactionHash string    `json:"transaction_hash"`
	ChainID         int       `json:"chain_id"`
	Status          string    `json:"status"`
	FromAddress     *string   `json:"from_address,omitempty"`
	ToAddress       *string   `json:"to_address,omitempty"`
	Value           *string   `json:"value,omitempty"`
	BlockNumber     *int64    `json:"block_number,omitempty"`
	GasUsed         *int64    `json:"gas_used,omitempty"`
	ErrorReason     *string   `json:"error_reason,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (h *Handlers) GetTransaction(w http.ResponseWriter, r *http.Request) {
	// ---- Extract hash from URL ----
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		http.Error(w, "transaction hash is required", http.StatusBadRequest)
		return
	}

	// ---- Request-scoped context with timeout ----
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// ---- Query transaction by hash ----
	query := `
		SELECT 
			id, 
			transaction_hash, 
			chain_id, 
			status, 
			from_address, 
			to_address, 
			value, 
			block_number, 
			gas_used, 
			error_reason, 
			created_at, 
			updated_at
		FROM transactions
		WHERE transaction_hash = $1
		LIMIT 1
	`

	var txn Transaction
	err := h.DB.QueryRow(ctx, query, hash).Scan(
		&txn.ID,
		&txn.TransactionHash,
		&txn.ChainID,
		&txn.Status,
		&txn.FromAddress,
		&txn.ToAddress,
		&txn.Value,
		&txn.BlockNumber,
		&txn.GasUsed,
		&txn.ErrorReason,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	)

	// ---- Handle errors ----
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// ---- Return transaction as JSON ----
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(txn)
}