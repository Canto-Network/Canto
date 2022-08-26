package types

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

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
	// smart contract address + NFT id -> last withdraw period
	prefixLastPeriodNFT
)

// KVStore key prefixes
var (
	KeyPrefixCSR           = []byte{prefixCSR}
	KeyPrefixSmartContract = []byte{prefixSmartContract}
	KeyPrefixDeployer      = []byte{prefixDeployer}
	KeyPrefixLastPeriodNFT = []byte{prefixLastPeriodNFT}
)

// Multi-tier index to find the latest withdrawal period for an NFT
func GetKeyPrefixLastPeriodNFT(contractAddress common.Address, id uint64) []byte {
	contractPrefix := append(KeyPrefixLastPeriodNFT, contractAddress.Bytes()...)
	return append(contractPrefix, []byte(strconv.FormatUint(id, 10))...)
}
