package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/csr module sentinel errors
var (
	// error in allocation of NFT supply in MsgRegisterCSR
	ErrMisMatchedAllocations = sdkerrors.Register(ModuleName, 6969, "csr::MsgRegisterCSR: MismatchedAllocations")
	ErrInvalidNFTSupply = sdkerrors.Register(ModuleName, 6970, "csr::MsgRegisterCSR: InvalidNFTSupply")
	ErrInvalidNonce = sdkerrors.Register(ModuleName, 6971, "csr::MsgRegisterCSR: InvalidNonce")
	ErrInvalidArity = sdkerrors.Register(ModuleName, 6972, "csr::MsgRegisterCSR: InvalidArity")
	ErrSmartContractSupply = sdkerrors.Register(ModuleName, 1000, "csr::CSRPool")
	ErrNFTSupply           = sdkerrors.Register(ModuleName, 1001, "csr::CSRPool")
	ErrMisMatchedNFTSupply = sdkerrors.Register(ModuleName, 1002, "csr::CSRPool")
	ErrInvalidParams       = sdkerrors.Register(ModuleName, 1003, "csr::Params")
	ErrDuplicatePools      = sdkerrors.Register(ModuleName, 1004, "csr::GenesisState")
	ErrDuplicateNFTs       = sdkerrors.Register(ModuleName, 1005, "csr::GenesisState")
)
