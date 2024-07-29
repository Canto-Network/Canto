package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Canto-Network/Canto/v8/x/erc20/types"
)

// GetTokenPairs - get all registered token tokenPairs
func (k Keeper) GetTokenPairs(ctx sdk.Context) []types.TokenPair {
	tokenPairs := []types.TokenPair{}

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixTokenPair)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tokenPair types.TokenPair
		k.cdc.MustUnmarshal(iterator.Value(), &tokenPair)

		tokenPairs = append(tokenPairs, tokenPair)
	}

	return tokenPairs
}

// GetTokenPairID returns the pair id from either of the registered tokens.
func (k Keeper) GetTokenPairID(ctx sdk.Context, token string) []byte {
	if common.IsHexAddress(token) {
		addr := common.HexToAddress(token)
		return k.GetTokenPairIdByERC20Addr(ctx, addr)
	}
	return k.GetTokenPairIdByDenom(ctx, token)
}

// GetTokenPair - get registered token pair from the identifier
func (k Keeper) GetTokenPair(ctx sdk.Context, id []byte) (types.TokenPair, bool) {
	if id == nil {
		return types.TokenPair{}, false
	}

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPair)
	var tokenPair types.TokenPair
	bz := prefixStore.Get(id)
	if len(bz) == 0 {
		return types.TokenPair{}, false
	}

	k.cdc.MustUnmarshal(bz, &tokenPair)
	return tokenPair, true
}

// SetTokenPair stores a token pair
func (k Keeper) SetTokenPair(ctx sdk.Context, tokenPair types.TokenPair) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPair)
	key := tokenPair.GetID()
	bz := k.cdc.MustMarshal(&tokenPair)
	prefixStore.Set(key, bz)
}

// DeleteTokenPair removes a token pair.
func (k Keeper) DeleteTokenPair(ctx sdk.Context, tokenPair types.TokenPair) {
	id := tokenPair.GetID()
	k.deleteTokenPair(ctx, id)
	k.deleteTokenPairIdByERC20Addr(ctx, tokenPair.GetERC20Contract())
	k.deleteTokenPairIdByDenom(ctx, tokenPair.Denom)
}

// deleteTokenPair deletes the token pair for the given id
func (k Keeper) deleteTokenPair(ctx sdk.Context, id []byte) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPair)
	prefixStore.Delete(id)
}

// GetTokenPairIdByERC20Addr returns the token pair id for the given address
func (k Keeper) GetTokenPairIdByERC20Addr(ctx sdk.Context, erc20 common.Address) []byte {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20Address)
	return prefixStore.Get(erc20.Bytes())
}

// GetTokenPairIdByDenom returns the token pair id for the given denomination
func (k Keeper) GetTokenPairIdByDenom(ctx sdk.Context, denom string) []byte {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	return prefixStore.Get([]byte(denom))
}

// SetTokenPairIdByERC20Addr sets the token pair id for the given address
func (k Keeper) SetTokenPairIdByERC20Addr(ctx sdk.Context, erc20 common.Address, id []byte) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20Address)
	prefixStore.Set(erc20.Bytes(), id)
}

// deleteTokenPairIdByERC20Addr deletes the token pair id for the given address
func (k Keeper) deleteTokenPairIdByERC20Addr(ctx sdk.Context, erc20 common.Address) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20Address)
	prefixStore.Delete(erc20.Bytes())
}

// SetTokenPairIdByDenom sets the token pair id for the denomination
func (k Keeper) SetTokenPairIdByDenom(ctx sdk.Context, denom string, id []byte) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	prefixStore.Set([]byte(denom), id)
}

// deleteTokenPairIdByDenom deletes the token pair id for the given denom
func (k Keeper) deleteTokenPairIdByDenom(ctx sdk.Context, denom string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	prefixStore.Delete([]byte(denom))
}

// IsTokenPairRegistered - check if registered token tokenPair is registered
func (k Keeper) IsTokenPairRegistered(ctx sdk.Context, id []byte) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPair)
	return prefixStore.Has(id)
}

// IsERC20Registered check if registered ERC20 token is registered
func (k Keeper) IsERC20Registered(ctx sdk.Context, erc20 common.Address) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20Address)
	return prefixStore.Has(erc20.Bytes())
}

// IsDenomRegistered check if registered coin denom is registered
func (k Keeper) IsDenomRegistered(ctx sdk.Context, denom string) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	return prefixStore.Has([]byte(denom))
}

// GetAllTokenPairDenomIndexes returns all token pair denom indexes
func (k Keeper) GetAllTokenPairDenomIndexes(ctx sdk.Context) []types.TokenPairDenomIndex {
	var idxs []types.TokenPairDenomIndex
	k.IterateTokenPairDenomIndex(ctx, func(denom string, id []byte) (stop bool) {
		idx := types.TokenPairDenomIndex{
			Denom:       denom,
			TokenPairId: id,
		}
		idxs = append(idxs, idx)
		return false
	})
	return idxs
}

// IterateTokenPairDenomIndex iterates over all token pair denom indexes
func (k Keeper) IterateTokenPairDenomIndex(ctx sdk.Context, cb func(denom string, id []byte) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		denom := string(iter.Key())
		id := iter.Value()
		if cb(denom, id) {
			break
		}
	}
}

// GetAllTokenPairERC20AddressIndexes returns all token pair ERC20 address indexes
func (k Keeper) GetAllTokenPairERC20AddressIndexes(ctx sdk.Context) []types.TokenPairERC20AddressIndex {
	var idxs []types.TokenPairERC20AddressIndex
	k.IterateTokenPairERC20AddressIndex(ctx, func(erc20Addr common.Address, id []byte) (stop bool) {
		idx := types.TokenPairERC20AddressIndex{
			Erc20Address: erc20Addr.Bytes(),
			TokenPairId:  id,
		}
		idxs = append(idxs, idx)
		return false
	})
	return idxs
}

// IterateTokenPairERC20AddressIndex iterates over all token pair ERC20 address indexes
func (k Keeper) IterateTokenPairERC20AddressIndex(ctx sdk.Context, cb func(erc20Addr common.Address, id []byte) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20Address)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		erc20Addr := common.BytesToAddress(iter.Key())
		id := iter.Value()
		if cb(erc20Addr, id) {
			break
		}
	}
}
