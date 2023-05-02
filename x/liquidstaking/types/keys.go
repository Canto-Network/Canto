package types

import (
	"bytes"

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
	prefixPairingInsuranceIndex
	prefixInsurancesByProviderIndex
	prefixWithdrawInsuranceRequest
	prefixPreviousInsuranceIndex
	prefixUnpairingForUnstakeChunkInfo
	prefixLiquidUnstakeKey
	prefixEpoch
)

// KVStore key prefixes
var (
	KeyPrefixLastChunkId                  = []byte{prefixLastChunkId}
	KeyPrefixLastInsuranceId              = []byte{prefixLastInsuranceId}
	KeyPrefixChunk                        = []byte{prefixChunk}
	KeyPrefixInsurance                    = []byte{prefixInsurance}
	KeyPrefixPairingInsuranceIndex        = []byte{prefixPairingInsuranceIndex}
	KeyPrefixInsurancesByProviderIndex    = []byte{prefixInsurancesByProviderIndex}
	KeyPrefixWithdrawInsuranceRequest     = []byte{prefixWithdrawInsuranceRequest}
	KeyPrefixUnpairingForUnstakeChunkInfo = []byte{prefixUnpairingForUnstakeChunkInfo}
	KeyPrefixLiquidUnstakeKey             = []byte{prefixLiquidUnstakeKey}
	KeyPrefixEpoch                        = []byte{prefixEpoch}
	KeyLiquidBondDenom                    = []byte{prefixLiquidBondDenom}
)

func GetChunkKey(chunkId uint64) []byte {
	return append(KeyPrefixChunk, sdk.Uint64ToBigEndian(chunkId)...)
}

func GetInsuranceKey(insuranceId uint64) []byte {
	return append(KeyPrefixInsurance, sdk.Uint64ToBigEndian(insuranceId)...)
}

func GetPairingInsuranceIndexKey(insuranceId uint64) []byte {
	return append(KeyPrefixPairingInsuranceIndex, sdk.Uint64ToBigEndian(insuranceId)...)
}

func GetInsurancesByProviderIndexKey(providerAddress sdk.AccAddress, insuranceId uint64) []byte {
	return append(append(KeyPrefixInsurancesByProviderIndex, address.MustLengthPrefix(providerAddress)...), sdk.Uint64ToBigEndian(insuranceId)...)
}

func GetWithdrawInsuranceRequestKey(insuranceId uint64) []byte {
	return append(KeyPrefixWithdrawInsuranceRequest, sdk.Uint64ToBigEndian(insuranceId)...)
}

func GetUnpairingForUnstakeChunkInfoKey(chunkId uint64) []byte {
	return append(KeyPrefixUnpairingForUnstakeChunkInfo, sdk.Uint64ToBigEndian(chunkId)...)
}

func ParseInsurancesByProviderIndexKey(key []byte) (providerAddress sdk.AccAddress, insuranceId uint64) {
	if !bytes.HasPrefix(key, KeyPrefixInsurancesByProviderIndex) {
		panic("invalid insurances by provider index key")
	}

	providerAddressLength := key[1]
	providerAddress = key[2 : 2+providerAddressLength]
	insuranceId = sdk.BigEndianToUint64(key[2+providerAddressLength:])
	return
}

func ParsePairingInsuranceIndexKey(key []byte) (insuranceId uint64) {
	if !bytes.HasPrefix(key, KeyPrefixPairingInsuranceIndex) {
		panic("invalid pairing insurance index key")
	}

	insuranceId = sdk.BigEndianToUint64(key[1:])
	return
}

func GetPendingLiquidStakeKey(delegator sdk.AccAddress) []byte {
	return append(KeyPrefixLiquidUnstakeKey, address.MustLengthPrefix(delegator)...)
}
