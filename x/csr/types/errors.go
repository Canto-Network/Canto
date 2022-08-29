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

	ErrMisMatchedAllocations = sdkerrors.Register(ModuleName, 6969, "csr::MsgRegisterCSR")
	ErrInvalidNFTSupply      = sdkerrors.Register(ModuleName, 6970, "csr::MsgRegisterCSR")
	ErrInvalidNonce          = sdkerrors.Register(ModuleName, 6971, "csr::MsgRegisterCSR")
	ErrInvalidArity          = sdkerrors.Register(ModuleName, 6972, "csr::MsgRegisterCSR")
	ErrInvalidType           = sdkerrors.Register(ModuleName, 6973, "csr::MsgRegisterCSR")
	ErrRepeatedNFT           = sdkerrors.Register(ModuleName, 6974, "csr::MsgWithdrawCSR")
)
