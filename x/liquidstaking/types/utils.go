package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

// DeriveAddress derives an address with the given address length type, module name, and
func DeriveAddress(moduleName, name string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(moduleName + name)))
}
