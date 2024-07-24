package types

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

// constants
const (
	// module name
	ModuleName = "erc20"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// ModuleAddress is the native module address for EVM
var ModuleAddress common.Address

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

// prefix bytes for the EVM persistent store
const (
	prefixTokenPair = iota + 1
	prefixTokenPairByERC20Address
	prefixTokenPairByDenom
)

// KVStore key prefixes
var (
	KeyPrefixTokenPair               = []byte{prefixTokenPair}
	KeyPrefixTokenPairByERC20Address = []byte{prefixTokenPairByERC20Address}
	KeyPrefixTokenPairByDenom        = []byte{prefixTokenPairByDenom}
)
