package blockchain

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func HexToInt64(hexStr string) (int64, error) {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")
	
	if hexStr == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hex: %w", err)
	}

	return value, nil
}

func HexToBigInt(hexStr string) (*big.Int, error) {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")
	
	if hexStr == "" {
		return big.NewInt(0), nil
	}

	value := new(big.Int)
	value, ok := value.SetString(hexStr, 16)
	if !ok {
		return nil, fmt.Errorf("failed to parse hex to big.Int")
	}

	return value, nil
}

func HexToDecimalString(hexStr string) (string, error) {
	bigInt, err := HexToBigInt(hexStr)
	if err != nil {
		return "", err
	}
	return bigInt.String(), nil
}
