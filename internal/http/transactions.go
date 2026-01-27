package models

import "time"

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
