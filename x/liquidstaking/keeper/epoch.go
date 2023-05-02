package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetEpoch(ctx sdk.Context) types.Epoch {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixEpoch)
	var epoch types.Epoch
	k.cdc.MustUnmarshal(bz, &epoch)
	return epoch
}
