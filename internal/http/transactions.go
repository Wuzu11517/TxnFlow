package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	var req createTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TransactionHash == "" || req.ChainID == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	//idempotent insert
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

	err := h.DB.QueryRow(
		ctx,
		insertQuery,
		req.TransactionHash,
		req.ChainID,
	).Scan(&id, &status, &createdAt, &updatedAt)

	//fetch existing row if insert didn't succeed
	if err != nil {
		selectQuery := `
			SELECT id, status, created_at, updated_at
			FROM transactions
			WHERE transaction_hash = $1 AND chain_id = $2
		`

		row := h.DB.QueryRow(
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

	//store ingestion event
	_, _ = h.DB.Exec(
		ctx,
		`INSERT INTO ingestion_events (transaction_id, new_status, reason)
		 VALUES ($1, $2, $3)`,
		id,
		status,
		"transaction registered",
	)

	//build response
	resp := map[string]interface{}{
		"id":               id,
		"transaction_hash": req.TransactionHash,
		"chain_id":         req.ChainID,
		"status":           status,
		"created_at":       createdAt,
	}

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
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		http.Error(w, "transaction hash is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

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

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(txn)
}

func (h *Handlers) ListTransactions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filters := struct {
		FromAddress    string
		ToAddress      string
		ChainID        string
		Status         string
		BlockNumberMin string
		BlockNumberMax string
		Limit          string
		Offset         string
	}{
		FromAddress:    query.Get("from_address"),
		ToAddress:      query.Get("to_address"),
		ChainID:        query.Get("chain_id"),
		Status:         query.Get("status"),
		BlockNumberMin: query.Get("block_number_min"),
		BlockNumberMax: query.Get("block_number_max"),
		Limit:          query.Get("limit"),
		Offset:         query.Get("offset"),
	}

	//default pagination values
	limit := 100
	offset := 0

	if filters.Limit != "" {
		if parsed, err := strconv.Atoi(filters.Limit); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	if filters.Offset != "" {
		if parsed, err := strconv.Atoi(filters.Offset); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	baseQuery := `
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
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}
	argCounter := 1

	if filters.FromAddress != "" {
		conditions = append(conditions, fmt.Sprintf("AND from_address = $%d", argCounter))
		args = append(args, filters.FromAddress)
		argCounter++
	}

	if filters.ToAddress != "" {
		conditions = append(conditions, fmt.Sprintf("AND to_address = $%d", argCounter))
		args = append(args, filters.ToAddress)
		argCounter++
	}

	if filters.ChainID != "" {
		if chainID, err := strconv.Atoi(filters.ChainID); err == nil {
			conditions = append(conditions, fmt.Sprintf("AND chain_id = $%d", argCounter))
			args = append(args, chainID)
			argCounter++
		}
	}

	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("AND status = $%d", argCounter))
		args = append(args, filters.Status)
		argCounter++
	}

	if filters.BlockNumberMin != "" {
		if blockNum, err := strconv.ParseInt(filters.BlockNumberMin, 10, 64); err == nil {
			conditions = append(conditions, fmt.Sprintf("AND block_number >= $%d", argCounter))
			args = append(args, blockNum)
			argCounter++
		}
	}

	if filters.BlockNumberMax != "" {
		if blockNum, err := strconv.ParseInt(filters.BlockNumberMax, 10, 64); err == nil {
			conditions = append(conditions, fmt.Sprintf("AND block_number <= $%d", argCounter))
			args = append(args, blockNum)
			argCounter++
		}
	}

	fullQuery := baseQuery + " " + strings.Join(conditions, " ") + 
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx, fullQuery, args...)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var txn Transaction
		err := rows.Scan(
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
		if err != nil {
			http.Error(w, "failed to scan result", http.StatusInternalServerError)
			return
		}
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "error reading results", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":   transactions,
		"limit":  limit,
		"offset": offset,
		"count":  len(transactions),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Count transactions by status
	query := `
		SELECT status, COUNT(*) 
		FROM transactions 
		GROUP BY status
	`

	rows, err := h.DB.Query(ctx, query)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	stats := make(map[string]int)
	totalCount := 0

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			http.Error(w, "failed to scan result", http.StatusInternalServerError)
			return
		}
		stats[status] = count
		totalCount += count
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "error reading results", http.StatusInternalServerError)
		return
	}

	// Build response
	response := map[string]interface{}{
		"total":        totalCount,
		"by_status":    stats,
		"timestamp":    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}