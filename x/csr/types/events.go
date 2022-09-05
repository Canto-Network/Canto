package types

import (
	"github.com/ethereum/go-ethereum/common"
)

const (
	TurnstileEventRegisterCSR         = "RegisterCSREvent"
	TurnstileEventUpdateCSR           = "UpdateCSREvent"
	TurnstileEventRetroactiveRegister = "RetroactiveRegisterEvent"
)

type RegisterCSREvent struct {
	SmartContractAddress common.Address
	Receiver             common.Address
}

type UpdateCSREvent struct {
	SmartContractAddress common.Address
	Id                   uint64
}
