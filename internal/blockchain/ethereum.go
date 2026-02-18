package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RPCClient struct {
	RPCURL     string
	HTTPClient *http.Client
}

// NewRPCClient creates a new Ethereum RPC client
func NewRPCClient(rpcURL string) *RPCClient {
	return &RPCClient{
		RPCURL: rpcURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

type EthTransaction struct {
	Hash             string `json:"hash"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex string `json:"transactionIndex"`
}

// EthTransactionReceipt represents a transaction receipt from eth_getTransactionReceipt
type EthTransactionReceipt struct {
	TransactionHash   string `json:"transactionHash"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Status            string `json:"status"`
}

// GetTransactionByHash fetches a transaction by hash
func (c *RPCClient) GetTransactionByHash(ctx context.Context, txHash string) (*EthTransaction, error) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionByHash",
		Params:  []interface{}{txHash},
		ID:      1,
	}

	var response JSONRPCResponse
	if err := c.call(ctx, request, &response); err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	// Check if result is null (transaction not found)
	if string(response.Result) == "null" {
		return nil, fmt.Errorf("transaction not found")
	}

	var tx EthTransaction
	if err := json.Unmarshal(response.Result, &tx); err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	return &tx, nil
}

// GetTransactionReceipt fetches a transaction receipt
func (c *RPCClient) GetTransactionReceipt(ctx context.Context, txHash string) (*EthTransactionReceipt, error) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionReceipt",
		Params:  []interface{}{txHash},
		ID:      1,
	}

	var response JSONRPCResponse
	if err := c.call(ctx, request, &response); err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	// Check if result is null (receipt not found - transaction might be pending)
	if string(response.Result) == "null" {
		return nil, fmt.Errorf("receipt not found (transaction may be pending)")
	}

	var receipt EthTransactionReceipt
	if err := json.Unmarshal(response.Result, &receipt); err != nil {
		return nil, fmt.Errorf("failed to parse receipt: %w", err)
	}

	return &receipt, nil
}

// call makes a JSON-RPC call to the Ethereum node
func (c *RPCClient) call(ctx context.Context, request JSONRPCRequest, response *JSONRPCResponse) error {
	// Serialize request
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.RPCURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RPC returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
