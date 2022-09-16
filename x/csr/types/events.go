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
	SmartContractAddress common.Address
	Receiver             common.Address
	Id                   *big.Int
}

// Update event that is emitted from the Turnstile Smart Contract
type UpdateCSREvent struct {
	SmartContractAddress common.Address
	Id                   *big.Int
}
