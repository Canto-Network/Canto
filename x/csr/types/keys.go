package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName defines the module name
	ModuleName = "csr"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName
)

const (
	// csr pool address -> csr struct
	prefixCSR = iota + 1
	// smartContract -> csr pool address
	prefixSmartContract
	// deployer address -> csr pool addresses
	prefixDeployer
	// csr pool address -> current period
	prefixCurrentPeriod
	// csr pool address -> cumulative rewards
	prefixCumulativeRewards
)

// KVStore key prefixes
var (
	KeyPrefixCSR               = []byte{prefixCSR}
	KeyPrefixSmartContract     = []byte{prefixSmartContract}
	KeyPrefixDeployer          = []byte{prefixDeployer}
	KeyPrefixCurrentPeriod     = []byte{prefixCurrentPeriod}
	KeyPrefixCumulativeRewards = []byte{prefixCumulativeRewards}
)

// Get the key for the pool so that we can extract rewards at a given period
func GetKeyPrefixPoolRewards(poolAddress sdk.AccAddress) []byte {
	return append(KeyPrefixCumulativeRewards, poolAddress.Bytes()...)
}
