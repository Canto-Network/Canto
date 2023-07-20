package keeper

import (
	"fmt"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetUnpairingForUnstakingChunkInfo(ctx sdk.Context, info types.UnpairingForUnstakingChunkInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&info)
	store.Set(types.GetUnpairingForUnstakingChunkInfoKey(info.ChunkId), bz)
}

func (k Keeper) GetUnpairingForUnstakingChunkInfo(ctx sdk.Context, id uint64) (info types.UnpairingForUnstakingChunkInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetUnpairingForUnstakingChunkInfoKey(id))
	if bz == nil {
		return info, false
	}
	k.cdc.MustUnmarshal(bz, &info)
	return info, true
}

func (k Keeper) mustGetUnpairingForUnstakingChunkInfo(ctx sdk.Context, id uint64) types.UnpairingForUnstakingChunkInfo {
	info, found := k.GetUnpairingForUnstakingChunkInfo(ctx, id)
	if !found {
		panic(fmt.Sprintf("unpairing for unstaking chunk info not found: %d", id))
	}
	return info
}

func (k Keeper) DeleteUnpairingForUnstakingChunkInfo(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetUnpairingForUnstakingChunkInfoKey(id))
}

func (k Keeper) IterateAllUnpairingForUnstakingChunkInfos(ctx sdk.Context, cb func(info types.UnpairingForUnstakingChunkInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixUnpairingForUnstakingChunkInfo)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info types.UnpairingForUnstakingChunkInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)

		stop := cb(info)
		if stop {
			break
		}
	}
}

func (k Keeper) GetAllUnpairingForUnstakingChunkInfos(ctx sdk.Context) (infos []types.UnpairingForUnstakingChunkInfo) {
	infos = []types.UnpairingForUnstakingChunkInfo{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixUnpairingForUnstakingChunkInfo)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info types.UnpairingForUnstakingChunkInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)

		infos = append(infos, info)
	}

	return infos
}
