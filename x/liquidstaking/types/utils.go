package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"math/big"
	"math/rand"
)

// DeriveAddress derives an address with the given address length type, module name, and
func DeriveAddress(moduleName, name string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(moduleName + name)))
}

// RandomInt returns a random integer in the half-open interval [min, max).
func RandomInt(r *rand.Rand, min, max sdk.Int) sdk.Int {
	return min.Add(sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.Sub(min).BigInt())))
}

// RandomDec returns a random decimal in the half-open interval [min, max).
func RandomDec(r *rand.Rand, min, max sdk.Dec) sdk.Dec {
	return min.Add(sdk.NewDecFromBigIntWithPrec(new(big.Int).Rand(r, max.Sub(min).BigInt()), sdk.Precision))
}
