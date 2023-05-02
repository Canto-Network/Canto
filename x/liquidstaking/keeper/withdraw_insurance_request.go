package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetWithdrawInsuranceRequest(ctx sdk.Context, req types.WithdrawInsuranceRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&req)
	store.Set(types.GetWithdrawInsuranceRequestKey(req.InsuranceId), bz)
}

func (k Keeper) GetWithdrawInsuranceRequest(ctx sdk.Context, id uint64) (req types.WithdrawInsuranceRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetWithdrawInsuranceRequestKey(id))
	if bz == nil {
		return req, false
	}
	k.cdc.MustUnmarshal(bz, &req)
	return req, true
}

func (k Keeper) DeleteWithdrawInsuranceRequest(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetWithdrawInsuranceRequestKey(id))
}

func (k Keeper) IterateWithdrawInsuranceRequests(ctx sdk.Context, cb func(req types.WithdrawInsuranceRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixWithdrawInsuranceRequest)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.WithdrawInsuranceRequest
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(req) {
			break
		}
	}
}

func (k Keeper) GetAllWithdrawInsuranceRequests(ctx sdk.Context) []types.WithdrawInsuranceRequest {
	var reqs []types.WithdrawInsuranceRequest

	k.IterateWithdrawInsuranceRequests(ctx, func(req types.WithdrawInsuranceRequest) bool {
		reqs = append(reqs, req)
		return false
	})

	return reqs
}
