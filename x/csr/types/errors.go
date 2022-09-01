package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/csr module sentinel errors
var (
	ErrSmartContractSupply         = sdkerrors.Register(ModuleName, 1000, "csr::CSR")
	ErrDuplicateSmartContracts     = sdkerrors.Register(ModuleName, 1001, "csr::CSR")
	ErrInvalidSmartContractAddress = sdkerrors.Register(ModuleName, 1002, "csr::CSR")

	ErrInvalidParams = sdkerrors.Register(ModuleName, 1003, "csr::Params")

	// keeper errors
	ErrAddressDerivation           = sdkerrors.Register(ModuleName, 1004, "csr::Keeper")
)