package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/csr module sentinel errors
var (
	ErrSmartContractSupply     = sdkerrors.Register(ModuleName, 1000, "csr::CSRPool")
	ErrNFTSupply               = sdkerrors.Register(ModuleName, 1001, "csr::CSRPool")
	ErrMisMatchedNFTSupply     = sdkerrors.Register(ModuleName, 1002, "csr::CSRPool")
	ErrInvalidParams           = sdkerrors.Register(ModuleName, 1003, "csr::Params")
	ErrDuplicatePools          = sdkerrors.Register(ModuleName, 1004, "csr::GenesisState")
	ErrDuplicateNFTs           = sdkerrors.Register(ModuleName, 1005, "csr::GenesisState")
	ErrDuplicateSmartContracts = sdkerrors.Register(ModuleName, 1006, "csr::CSR")
  ErrMisMatchedAllocations   = sdkerrors.Register(ModuleName, 6969, "csr::MsgRegisterCSR")
	ErrInvalidNFTSupply        = sdkerrors.Register(ModuleName, 6970, "csr::MsgRegisterCSR")
	ErrInvalidNonce            = sdkerrors.Register(ModuleName, 6971, "csr::MsgRegisterCSR")
	ErrInvalidArity            = sdkerrors.Register(ModuleName, 6972, "csr::MsgRegisterCSR")
	ErrInvalidType             = sdkerrors.Register(ModuleName, 6973, "csr::MsgRegisterCSR")
  ErrRepeatedNFT             = sdkerrors.Register(ModuleName, 6974, "csr::MsgWithdrawCSR")
)
