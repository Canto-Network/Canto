package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	TurnstileEventRegister = "Register"
	TurnstileEventUpdate   = "Assign"
)

// Register Event that is emitted from the Turnstile Smart Contract
type RegisterCSREvent struct {
	SmartContract common.Address
	Recipient     common.Address
	TokenId       *big.Int
}

// Update event that is emitted from the Turnstile Smart Contract
type UpdateCSREvent struct {
	SmartContract common.Address
	TokenId       *big.Int
}
