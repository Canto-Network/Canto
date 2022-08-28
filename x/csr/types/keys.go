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
	// csr pool address -> deployer
	prefixCSRPoolDeployer = iota + 1
	// csr pool address -> contracts
	prefixCSRPoolContracts
	// csr pool address -> nfts
	prefixCSRPoolNFTs
	// csr pool address
	prefixCSRPoolNFTSupply

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
	KeyPrefixCSRPoolDeployer  = []byte{prefixCSRPoolDeployer}
	KeyPrefixCSRPoolContracts = []byte{prefixCSRPoolContracts}
	KeyPrefixCSRPoolNFTs      = []byte{prefixCSRPoolNFTs}
	KeyPrefixCSRPoolNFTSupply = []byte{prefixCSRPoolNFTSupply}

	KeyPrefixDeployer          = []byte{prefixDeployer}
	KeyPrefixCurrentPeriod     = []byte{prefixCurrentPeriod}
	KeyPrefixCumulativeRewards = []byte{prefixCumulativeRewards}
)

// Get the key for the pool so that we can extract rewards at a given period
func GetKeyPrefixPoolRewards(poolAddress sdk.AccAddress) []byte {
	return append(KeyPrefixCumulativeRewards, poolAddress.Bytes()...)
}

// Get the key for the set of all contracts for a pool, this will be indexed via contract number
func GetKeyPrefixPoolContracts(poolAddress sdk.AccAddress) []byte {
	return append(KeyPrefixCSRPoolContracts, poolAddress.Bytes()...)
}
