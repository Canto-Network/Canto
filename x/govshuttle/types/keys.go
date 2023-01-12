package types

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ModuleName defines the module name
	ModuleName = "govshuttle"

	// StoreKey defines the primary module store key
	StoreKey = "shuttle"

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_govshuttle"
)

var (
	ModuleAddress common.Address
	PortKey       = []byte("Port")
)

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
