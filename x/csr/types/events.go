package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	TurnstileEventRegister = "Register"
	TurnstileEventUpdate   = "Attach"
)

type RegisterCSREvent struct {
	SmartContractAddress common.Address
	Receiver             common.Address
	Id                   *big.Int
}

type UpdateCSREvent struct {
	SmartContractAddress common.Address
	Id                   *big.Int
}
