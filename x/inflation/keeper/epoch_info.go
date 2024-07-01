package keeper

import (
	"github.com/Canto-Network/Canto/v7/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEpochIdentifier gets the epoch identifier
func (k Keeper) GetEpochIdentifier(ctx sdk.Context) string {
	store := k.storeService.OpenKVStore(ctx)
	bz, _ := store.Get(types.KeyPrefixEpochIdentifier)
	if len(bz) == 0 {
		return ""
	}

	return string(bz)
}

// SetEpochsPerPeriod stores the epoch identifier
func (k Keeper) SetEpochIdentifier(ctx sdk.Context, epochIdentifier string) {
	store := k.storeService.OpenKVStore(ctx)
	store.Set(types.KeyPrefixEpochIdentifier, []byte(epochIdentifier))
}

// GetEpochsPerPeriod gets the epochs per period
func (k Keeper) GetEpochsPerPeriod(ctx sdk.Context) int64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, _ := store.Get(types.KeyPrefixEpochsPerPeriod)
	if len(bz) == 0 {
		return 0
	}

	return int64(sdk.BigEndianToUint64(bz))
}

// SetEpochsPerPeriod stores the epochs per period
func (k Keeper) SetEpochsPerPeriod(ctx sdk.Context, epochsPerPeriod int64) {
	store := k.storeService.OpenKVStore(ctx)
	store.Set(types.KeyPrefixEpochsPerPeriod, sdk.Uint64ToBigEndian(uint64(epochsPerPeriod)))
}

// GetSkippedEpochs gets the number of skipped epochs
func (k Keeper) GetSkippedEpochs(ctx sdk.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, _ := store.Get(types.KeyPrefixSkippedEpochs)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetSkippedEpochs stores the number of skipped epochs
func (k Keeper) SetSkippedEpochs(ctx sdk.Context, skippedEpochs uint64) {
	store := k.storeService.OpenKVStore(ctx)
	store.Set(types.KeyPrefixSkippedEpochs, sdk.Uint64ToBigEndian(skippedEpochs))
}
