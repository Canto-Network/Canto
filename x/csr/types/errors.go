package types

import (
	errorsmod "cosmossdk.io/errors"
)

// CSR module sentinel errors
var (
	ErrSmartContractSupply         = errorsmod.Register(ModuleName, 1000, "The supply of smart contracts must be greater than 0")
	ErrDuplicateSmartContracts     = errorsmod.Register(ModuleName, 1001, "There cannot be duplicate smart contracts")
	ErrInvalidSmartContractAddress = errorsmod.Register(ModuleName, 1002, "There cannot be invalid smart contract addresses")
	ErrInvalidParams               = errorsmod.Register(ModuleName, 1003, "The parameters for CSR are invalid")
)
