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
	// ModuleAcct will be the account from which all contracts are deployed from
)

var ModuleAddress common.Address

// instantiate new new EVM Account for module, and register w authkeeper
func init() { 
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}


const (
	// nft id -> csr
	prefixCSR = iota + 1
	// nft id -> owner
	prefixOwner
	// contract -> nft id
	prefixContract
)

// KVStore key prefixes
var (
	KeyPrefixCSR      = []byte{prefixCSR}
	KeyPrefixOwner    = []byte{prefixOwner}
	KeyPrefixContract = []byte{prefixContract}
)
