package keeper

import (
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

func (k Keeper) DeleteUnpairingForUnstakingChunkInfo(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetUnpairingForUnstakingChunkInfoKey(id))
}

func (k Keeper) IterateAllUnpairingForUnstakingChunkInfos(ctx sdk.Context, cb func(info types.UnpairingForUnstakingChunkInfo) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixUnpairingForUnstakingChunkInfo)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info types.UnpairingForUnstakingChunkInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)

		stop, err := cb(info)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}

	return nil
}

func (k Keeper) GetAllUnpairingForUnstakingChunkInfos(ctx sdk.Context) []types.UnpairingForUnstakingChunkInfo {
	var infos []types.UnpairingForUnstakingChunkInfo

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
