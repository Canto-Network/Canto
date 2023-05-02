package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: Use key with chunk id
func (k Keeper) SetPendingLiquidUnstake(ctx sdk.Context, pendingLiquidUnstake types.PendingLiquidUnstake) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&pendingLiquidUnstake)
	store.Set(types.GetPendingLiquidStakeKey(pendingLiquidUnstake.Delegator()), bz)
}

func (k Keeper) GetAllPendingLiquidUnstake(ctx sdk.Context) []types.PendingLiquidUnstake {
	var pendingLiquidUnstakes []types.PendingLiquidUnstake
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixLiquidUnstakeKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pendingLiquidUnstake types.PendingLiquidUnstake
		k.cdc.MustUnmarshal(iterator.Value(), &pendingLiquidUnstake)
		pendingLiquidUnstakes = append(pendingLiquidUnstakes, pendingLiquidUnstake)
	}
	return pendingLiquidUnstakes
}

func (k Keeper) DeletePendingLiquidUnstake(ctx sdk.Context, pendingLiquidUnstake types.PendingLiquidUnstake) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPendingLiquidStakeKey(pendingLiquidUnstake.Delegator()))
}
