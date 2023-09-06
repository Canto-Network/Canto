package keeper

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetRedelegationInfo(ctx sdk.Context, info types.RedelegationInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&info)
	store.Set(types.GetRedelegationInfoKey(info.ChunkId), bz)
}

func (k Keeper) GetRedelegationInfo(ctx sdk.Context, id uint64) (info types.RedelegationInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRedelegationInfoKey(id))
	if bz == nil {
		return info, false
	}
	k.cdc.MustUnmarshal(bz, &info)
	return info, true
}

func (k Keeper) DeleteRedelegationInfo(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetRedelegationInfoKey(id))
}

func (k Keeper) IterateAllRedelegationInfos(ctx sdk.Context, cb func(info types.RedelegationInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixRedelegationInfo)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info types.RedelegationInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)

		stop := cb(info)
		if stop {
			break
		}
	}
}

func (k Keeper) GetAllRedelegationInfos(ctx sdk.Context) (infos []types.RedelegationInfo) {
	infos = []types.RedelegationInfo{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixRedelegationInfo)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info types.RedelegationInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)
		infos = append(infos, info)
	}

	return infos
}
