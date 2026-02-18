package blockchain

import (
	"fmt"
)

// ChainConfig holds configuration for a blockchain network
type ChainConfig struct {
	ChainID int
	Name    string
	RPCURL  string
	Type    ChainType
}

// ChainType represents the blockchain type
type ChainType string

const (
	ChainTypeEVM ChainType = "EVM" // Ethereum Virtual Machine compatible
)

// ChainRegistry manages blockchain configurations
type ChainRegistry struct {
	chains map[int]*ChainConfig
}

// NewChainRegistry creates a new chain registry with default configurations
func NewChainRegistry(infuraAPIKey string) *ChainRegistry {
	registry := &ChainRegistry{
		chains: make(map[int]*ChainConfig),
	}

	// Ethereum Mainnet
	registry.RegisterChain(&ChainConfig{
		ChainID: 1,
		Name:    "Ethereum Mainnet",
		RPCURL:  fmt.Sprintf("https://mainnet.infura.io/v3/%s", infuraAPIKey),
		Type:    ChainTypeEVM,
	})

	// Future: Add more chains here
	// Polygon
	// registry.RegisterChain(&ChainConfig{
	//     ChainID: 137,
	//     Name:    "Polygon",
	//     RPCURL:  "https://polygon-rpc.com",
	//     Type:    ChainTypeEVM,
	// })
	//
	// Arbitrum
	// registry.RegisterChain(&ChainConfig{
	//     ChainID: 42161,
	//     Name:    "Arbitrum One",
	//     RPCURL:  "https://arb1.arbitrum.io/rpc",
	//     Type:    ChainTypeEVM,
	// })

	return registry
}

// RegisterChain adds a chain configuration to the registry
func (r *ChainRegistry) RegisterChain(config *ChainConfig) {
	r.chains[config.ChainID] = config
}

// GetChain returns the configuration for a chain ID
func (r *ChainRegistry) GetChain(chainID int) (*ChainConfig, error) {
	config, ok := r.chains[chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}
	return config, nil
}

// IsSupported checks if a chain ID is supported
func (r *ChainRegistry) IsSupported(chainID int) bool {
	_, ok := r.chains[chainID]
	return ok
}

// GetSupportedChains returns a list of all supported chain IDs
func (r *ChainRegistry) GetSupportedChains() []int {
	chains := make([]int, 0, len(r.chains))
	for chainID := range r.chains {
		chains = append(chains, chainID)
	}
	return chains
}
