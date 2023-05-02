package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetUnpairingForUnstakeChunkInfo(ctx sdk.Context, info types.UnpairingForUnstakeChunkInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&info)
	store.Set(types.GetUnpairingForUnstakeChunkInfoKey(info.ChunkId), bz)
}

func (k Keeper) GetUnpairingForUnstakeChunkInfo(ctx sdk.Context, id uint64) (info types.UnpairingForUnstakeChunkInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetUnpairingForUnstakeChunkInfoKey(id))
	if bz == nil {
		return info, false
	}
	k.cdc.MustUnmarshal(bz, &info)
	return info, true
}

func (k Keeper) DeleteUnpairingForUnstakeChunkInfo(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetUnpairingForUnstakeChunkInfoKey(id))
}

func (k Keeper) GetAllUnpairingForUnstakeChunkInfos(ctx sdk.Context) []types.UnpairingForUnstakeChunkInfo {
	var infos []types.UnpairingForUnstakeChunkInfo

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixUnpairingForUnstakeChunkInfo)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info types.UnpairingForUnstakeChunkInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)

		infos = append(infos, info)
	}

	return infos
}
