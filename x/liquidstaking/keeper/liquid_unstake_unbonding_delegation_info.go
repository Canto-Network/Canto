package keeper

import (
	"encoding/binary"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetLiquidUnstakeUnbondingDelegationInfo(ctx sdk.Context, liquidUnstakeUnbondingDelegationInfo types.LiquidUnstakeUnbondingDelegationInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLiquidUnstakeUnbondingDelegation)
	chunkId := make([]byte, 8)
	binary.LittleEndian.PutUint64(chunkId, liquidUnstakeUnbondingDelegationInfo.ChunkId)
	bz := k.cdc.MustMarshal(&liquidUnstakeUnbondingDelegationInfo)
	store.Set(chunkId, bz)
}

func (k Keeper) GetLiquidUnstakeUnbondingDelegationInfo(ctx sdk.Context, chunkId uint64) (liquidUnstakeUnbondingDelegationInfo types.LiquidUnstakeUnbondingDelegationInfo, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLiquidUnstakeUnbondingDelegation)
	chunkIdBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(chunkIdBytes, chunkId)
	bz := store.Get(chunkIdBytes)
	if bz == nil {
		return liquidUnstakeUnbondingDelegationInfo, false
	}
	k.cdc.MustUnmarshal(bz, &liquidUnstakeUnbondingDelegationInfo)
	return liquidUnstakeUnbondingDelegationInfo, true
}

func (k Keeper) DeleteLiquidUnstakeUnbondingDelegationInfo(ctx sdk.Context, chunkId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLiquidUnstakeUnbondingDelegation)
	chunkIdBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(chunkIdBytes, chunkId)
	store.Delete(chunkIdBytes)
}

func (k Keeper) GetLiquidUnstakeUnbondingDelegationInfos(ctx sdk.Context) []types.LiquidUnstakeUnbondingDelegationInfo {
	var liquidUnstakeUnbondingDelegations []types.LiquidUnstakeUnbondingDelegationInfo

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixLiquidUnstakeUnbondingDelegation)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var liquidUnstakeUnbondingDelegationInfo types.LiquidUnstakeUnbondingDelegationInfo
		k.cdc.MustUnmarshal(iterator.Value(), &liquidUnstakeUnbondingDelegationInfo)

		liquidUnstakeUnbondingDelegations = append(liquidUnstakeUnbondingDelegations, liquidUnstakeUnbondingDelegationInfo)
	}

	return liquidUnstakeUnbondingDelegations
}
