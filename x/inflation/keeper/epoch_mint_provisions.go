package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v7/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEpochMintProvision gets the current EpochMintProvision
func (k Keeper) GetEpochMintProvision(ctx sdk.Context) (sdkmath.LegacyDec, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixEpochMintProvision)
	if len(bz) == 0 {
		return sdkmath.LegacyZeroDec(), false
	}

	var epochMintProvision sdkmath.LegacyDec
	err := epochMintProvision.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal epochMintProvision value: %w", err))
	}

	return epochMintProvision, true
}

// SetEpochMintProvision sets the current EpochMintProvision
func (k Keeper) SetEpochMintProvision(ctx sdk.Context, epochMintProvision sdkmath.LegacyDec) {
	bz, err := epochMintProvision.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal amount value: %w", err))
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixEpochMintProvision, bz)
}
