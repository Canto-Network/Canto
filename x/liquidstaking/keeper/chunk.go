package keeper

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"
)

func (k Keeper) SetChunk(ctx sdk.Context, chunk types.Chunk) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&chunk)
	store.Set(types.GetChunkKey(chunk.Id), bz)
}

func (k Keeper) GetChunk(ctx sdk.Context, id uint64) (chunk types.Chunk, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetChunkKey(id))
	if bz == nil {
		return chunk, false
	}
	k.cdc.MustUnmarshal(bz, &chunk)
	return chunk, true
}

func (k Keeper) mustGetChunk(ctx sdk.Context, id uint64) types.Chunk {
	chunk, found := k.GetChunk(ctx, id)
	if !found {
		panic("chunk not found")
	}
	return chunk
}

func (k Keeper) DeleteChunk(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetChunkKey(id))
}

func (k Keeper) GetAllPairingChunks(ctx sdk.Context) (chunks []types.Chunk) {
	k.IterateAllChunks(ctx, func(chunk types.Chunk) (stop bool) {
		if chunk.Status == types.CHUNK_STATUS_PAIRING {
			chunks = append(chunks, chunk)
		}
		return false
	})
	return
}

func (k Keeper) IterateAllChunks(ctx sdk.Context, cb func(chunk types.Chunk) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixChunk)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var chunk types.Chunk
		k.cdc.MustUnmarshal(iterator.Value(), &chunk)

		stop := cb(chunk)
		if stop {
			break
		}
	}
}

func (k Keeper) GetAllChunks(ctx sdk.Context) (chunks []types.Chunk) {
	chunks = []types.Chunk{}
	k.IterateAllChunks(ctx, func(chunk types.Chunk) (stop bool) {
		chunks = append(chunks, chunk)
		return false
	})
	return
}

func (k Keeper) SetLastChunkId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.KeyPrefixLastChunkId, bz)
}

func (k Keeper) GetLastChunkId(ctx sdk.Context) (id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixLastChunkId)
	if bz == nil {
		id = 0
	} else {
		var val gogotypes.UInt64Value
		k.cdc.MustUnmarshal(bz, &val)
		id = val.GetValue()
	}
	return
}

func (k Keeper) getNextChunkIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetLastChunkId(ctx) + 1
	k.SetLastChunkId(ctx, id)
	return id
}
