package keeper

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetEpoch(ctx sdk.Context) types.Epoch {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixEpoch)
	var epoch types.Epoch
	k.cdc.MustUnmarshal(bz, &epoch)
	return epoch
}

func (k Keeper) SetEpoch(ctx sdk.Context, epoch types.Epoch) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&epoch)
	store.Set(types.KeyPrefixEpoch, bz)
}

func (k Keeper) IncrementEpoch(ctx sdk.Context) {
	epoch := k.GetEpoch(ctx)
	epoch.CurrentNumber++
	epoch.StartTime = ctx.BlockTime()
	epoch.StartHeight = ctx.BlockHeight()
	k.SetEpoch(ctx, epoch)
}

func (k Keeper) IsEpochReached(ctx sdk.Context) bool {
	epoch := k.GetEpoch(ctx)
	return !ctx.BlockTime().Before(epoch.StartTime.Add(epoch.Duration))
}
