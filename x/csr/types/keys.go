package types

import (
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
	// smart contract address -> NFTs
	prefixNFT
)

// KVStore key prefixes
var (
	KeyPrefixCSR           = []byte{prefixCSR}
	KeyPrefixSmartContract = []byte{prefixSmartContract}
	KeyPrefixDeployer      = []byte{prefixDeployer}
	KeyPrefixNFT           = []byte{prefixNFT}
)

// Multi-tier index to find the latest withdrawal period for an NFT
func GetKeyPrefixLastPeriodNFT(contractAddress common.Address) []byte {
	return append(KeyPrefixNFT, contractAddress.Bytes()...)
}
