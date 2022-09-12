package types

import (
	"github.com/ethereum/go-ethereum/common"
)

const (
	TurnstileEventRegister = "Register"
	TurnstileEventUpdate   = "Attach"
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
