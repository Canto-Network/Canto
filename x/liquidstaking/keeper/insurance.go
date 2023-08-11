package keeper

import (
	"fmt"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogotypes "github.com/gogo/protobuf/types"
)

func (k Keeper) SetInsurance(ctx sdk.Context, ins types.Insurance) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&ins)
	store.Set(types.GetInsuranceKey(ins.Id), bz)
}

func (k Keeper) GetInsurance(ctx sdk.Context, id uint64) (ins types.Insurance, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetInsuranceKey(id))
	if bz == nil {
		return ins, false
	}
	k.cdc.MustUnmarshal(bz, &ins)
	return ins, true
}

func (k Keeper) mustGetInsurance(ctx sdk.Context, id uint64) types.Insurance {
	ins, found := k.GetInsurance(ctx, id)
	if !found {
		panic(fmt.Sprintf("insurance not found(id: %d)", id))
	}
	return ins
}

func (k Keeper) DeleteInsurance(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	ins, _ := k.GetInsurance(ctx, id)
	store.Delete(types.GetInsuranceKey(ins.Id))
}

func (k Keeper) GetPairingInsurances(ctx sdk.Context) (
	pairingInsurances []types.Insurance,
	validatorMap map[string]stakingtypes.Validator,
) {
	validatorMap = make(map[string]stakingtypes.Validator)
	k.IterateAllInsurances(ctx, func(ins types.Insurance) bool {
		if ins.Status != types.INSURANCE_STATUS_PAIRING {
			return false
		}
		if _, ok := validatorMap[ins.ValidatorAddress]; !ok {
			validator, found := k.stakingKeeper.GetValidator(ctx, ins.GetValidator())
			if !found {
				return false
			}
			if err := k.ValidateValidator(ctx, validator); err != nil {
				return false
			}
			validatorMap[ins.ValidatorAddress] = validator
		}
		pairingInsurances = append(pairingInsurances, ins)
		return false
	})
	return
}

func (k Keeper) IterateAllInsurances(ctx sdk.Context, cb func(ins types.Insurance) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixInsurance)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var ins types.Insurance
		k.cdc.MustUnmarshal(iterator.Value(), &ins)

		stop := cb(ins)
		if stop {
			break
		}
	}
}

func (k Keeper) GetAllInsurances(ctx sdk.Context) (inss []types.Insurance) {
	inss = []types.Insurance{}
	k.IterateAllInsurances(ctx, func(ins types.Insurance) (stop bool) {
		inss = append(inss, ins)
		return false
	})
	return
}

func (k Keeper) SetLastInsuranceId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.KeyPrefixLastInsuranceId, bz)
}

func (k Keeper) GetLastInsuranceId(ctx sdk.Context) (id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixLastInsuranceId)
	if bz == nil {
		id = 0
	} else {
		var val gogotypes.UInt64Value
		k.cdc.MustUnmarshal(bz, &val)
		id = val.GetValue()
	}
	return
}

func (k Keeper) getNextInsuranceIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetLastInsuranceId(ctx) + 1
	k.SetLastInsuranceId(ctx, id)
	return id
}
