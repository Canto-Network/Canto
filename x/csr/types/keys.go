package types

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

// ModuleAcct will be the account from which all contracts are deployed from
var ModuleAddress common.Address

// key for turnstile address once deployed
var TurnstileKey = []byte("Turnstile")

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

const (
	// nft id -> csr
	prefixCSR = iota + 1
	// contract -> nft id
	prefixContract
	// prefix address of the Turnstile smart contract
	prefixAddrs
)

// KVStore key prefixes
var (
	KeyPrefixCSR      = []byte{prefixCSR}
	KeyPrefixContract = []byte{prefixContract}
	KeyPrefixAddrs    = []byte{prefixAddrs}
)
