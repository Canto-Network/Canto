package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// constants
const (
	// ModuleName is the name of the module
	ModuleName = "liquidstaking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the liquidstaking module
	RouterKey = ModuleName
)

// prefix bytes for the liquidstaking persistent store
const (
	prefixLiquidBondDenom = iota + 1
	prefixLastChunkId
	prefixLastInsuranceId
	prefixChunk
	prefixInsurance
	prefixWithdrawInsuranceRequest
	prefixUnpairingForUnstakingChunkInfo
	prefixLiquidUnstakeKey
	prefixEpoch
)

// KVStore key prefixes
var (
	KeyPrefixLastChunkId                    = []byte{prefixLastChunkId}
	KeyPrefixLastInsuranceId                = []byte{prefixLastInsuranceId}
	KeyPrefixChunk                          = []byte{prefixChunk}
	KeyPrefixInsurance                      = []byte{prefixInsurance}
	KeyPrefixWithdrawInsuranceRequest       = []byte{prefixWithdrawInsuranceRequest}
	KeyPrefixUnpairingForUnstakingChunkInfo = []byte{prefixUnpairingForUnstakingChunkInfo}
	KeyPrefixLiquidUnstakeKey               = []byte{prefixLiquidUnstakeKey}
	KeyPrefixEpoch                          = []byte{prefixEpoch}
	KeyLiquidBondDenom                      = []byte{prefixLiquidBondDenom}
)

func GetChunkKey(chunkId uint64) []byte {
	return append(KeyPrefixChunk, sdk.Uint64ToBigEndian(chunkId)...)
}

func GetInsuranceKey(insuranceId uint64) []byte {
	return append(KeyPrefixInsurance, sdk.Uint64ToBigEndian(insuranceId)...)
}

func GetWithdrawInsuranceRequestKey(insuranceId uint64) []byte {
	return append(KeyPrefixWithdrawInsuranceRequest, sdk.Uint64ToBigEndian(insuranceId)...)
}

func GetUnpairingForUnstakingChunkInfoKey(chunkId uint64) []byte {
	return append(KeyPrefixUnpairingForUnstakingChunkInfo, sdk.Uint64ToBigEndian(chunkId)...)
}

func GetPendingLiquidStakeKey(delegator sdk.AccAddress) []byte {
	return append(KeyPrefixLiquidUnstakeKey, address.MustLengthPrefix(delegator)...)
}
