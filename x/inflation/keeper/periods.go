package keeper

import (
	"github.com/Canto-Network/Canto/v7/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPeriod gets current period
func (k Keeper) GetPeriod(ctx sdk.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, _ := store.Get(types.KeyPrefixPeriod)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetPeriod stores the current period
func (k Keeper) SetPeriod(ctx sdk.Context, period uint64) {
	store := k.storeService.OpenKVStore(ctx)
	store.Set(types.KeyPrefixPeriod, sdk.Uint64ToBigEndian(period))
}
