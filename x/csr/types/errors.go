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
)
