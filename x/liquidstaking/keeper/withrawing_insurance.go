package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetWithdrawingInsurance(ctx sdk.Context, withdrawingInsurance types.WithdrawingInsurance) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&withdrawingInsurance)
	store.Set(types.GetWithdrawingInsuranceKey(withdrawingInsurance.InsuranceId), bz)
}

func (k Keeper) GetWithdrawingInsurance(ctx sdk.Context, id uint64) (withdrawingInsurance types.WithdrawingInsurance, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetWithdrawingInsuranceKey(id))
	if bz == nil {
		return withdrawingInsurance, false
	}
	k.cdc.MustUnmarshal(bz, &withdrawingInsurance)
	return withdrawingInsurance, true
}

func (k Keeper) DeleteWithdrawingInsurance(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetWithdrawingInsuranceKey(id))
}

func (k Keeper) GetWithdrawingInsurances(ctx sdk.Context) []types.WithdrawingInsurance {
	var withdrawingInsurances []types.WithdrawingInsurance

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixWithdrawingInsurance)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var withdrawingInsurance types.WithdrawingInsurance
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawingInsurance)

		withdrawingInsurances = append(withdrawingInsurances, withdrawingInsurance)
	}

	return withdrawingInsurances
}
