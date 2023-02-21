package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CSR module sentinel errors
var (
	ErrSmartContractSupply         = sdkerrors.Register(ModuleName, 1000, "The supply of smart contracts must be greater than 0")
	ErrDuplicateSmartContracts     = sdkerrors.Register(ModuleName, 1001, "There cannot be duplicate smart contracts")
	ErrInvalidSmartContractAddress = sdkerrors.Register(ModuleName, 1002, "There cannot be invalid smart contract addresses")
	ErrInvalidParams               = sdkerrors.Register(ModuleName, 1003, "The parameters for CSR are invalid")
)
