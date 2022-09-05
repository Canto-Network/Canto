package types

import (
	"github.com/ethereum/go-ethereum/common"
)

const (
	TurnstileEventRegisterCSR         = "RegisterCSREvent"
	TurnstileEventUpdateCSR           = "UpdateCSREvent"
	TurnstileEventRetroactiveRegister = "RetroactiveRegisterEvent"
	WithdrawalEvent                   = "Withdrawal"
)

type RegisterCSREvent struct {
	SmartContractAddress common.Address
	Receiver             common.Address
}

type UpdateCSREvent struct {
	SmartContractAddress common.Address
	Nft_id               uint64
}

type Withdrawal struct {
	Withdrawer common.Address
	Receiver   common.Address
	Id         uint64
}
