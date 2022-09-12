package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	TurnstileEventRegister            = "Register"
	TurnstileEventUpdate              = "Attach"
	TurnstileEventRetroactiveRegister = "RetroactiveRegisterEvent"
	WithdrawalEvent                   = "Withdrawal"
)

type RegisterCSREvent struct {
	SmartContractAddress common.Address
	Receiver             common.Address
	Id                   uint64
}

type UpdateCSREvent struct {
	SmartContractAddress common.Address
	Id                   uint64
}

type Withdrawal struct {
	Withdrawer common.Address
	Receiver   common.Address
	Id         *big.Int
}
