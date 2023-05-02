package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogotypes "github.com/gogo/protobuf/types"
)

func (k Keeper) SetInsurance(ctx sdk.Context, insurance types.Insurance) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&insurance)
	store.Set(types.GetInsuranceKey(insurance.Id), bz)
}

func (k Keeper) GetPairingInsurance(ctx sdk.Context, id uint64) (insurance types.Insurance, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPairingInsuranceIndexKey(id))
	if bz == nil {
		return insurance, false
	}
	return k.GetInsurance(ctx, id)
}

func (k Keeper) SetPairingInsuranceIndex(ctx sdk.Context, insurance types.Insurance) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPairingInsuranceIndexKey(insurance.Id), []byte{})
}

func (k Keeper) DeletePairingInsuranceIndex(ctx sdk.Context, insurance types.Insurance) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPairingInsuranceIndexKey(insurance.Id))
}

func (k Keeper) SetInsurancesByProviderIndex(ctx sdk.Context, insurance types.Insurance) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetInsurancesByProviderIndexKey(insurance.GetProvider(), insurance.Id), []byte{})
}

func (k Keeper) DeleteInsurancesByProviderIndex(ctx sdk.Context, insurance types.Insurance) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetInsurancesByProviderIndexKey(insurance.GetProvider(), insurance.Id))
}

func (k Keeper) GetInsurance(ctx sdk.Context, id uint64) (insurance types.Insurance, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetInsuranceKey(id))
	if bz == nil {
		return insurance, false
	}
	k.cdc.MustUnmarshal(bz, &insurance)
	return insurance, true
}

func (k Keeper) DeleteInsurance(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	insurance, _ := k.GetInsurance(ctx, id)
	store.Delete(types.GetInsuranceKey(insurance.Id))
	store.Delete(types.GetInsurancesByProviderIndexKey(sdk.AccAddress(insurance.ProviderAddress), insurance.Id))
}

func (k Keeper) getPairingInsurances(ctx sdk.Context) (
	pairingInsurances []types.Insurance,
	validatorMap map[string]stakingtypes.Validator,
) {
	validatorMap = make(map[string]stakingtypes.Validator)
	err := k.IteratePairingInsurances(ctx, func(insurance types.Insurance) (bool, error) {
		if _, ok := validatorMap[insurance.ValidatorAddress]; !ok {
			validator, found := k.stakingKeeper.GetValidator(ctx, insurance.GetValidator())
			err := k.IsValidValidator(ctx, validator, found)
			if err != nil {
				return false, nil
			}
			validatorMap[insurance.ValidatorAddress] = validator
		}
		pairingInsurances = append(pairingInsurances, insurance)
		return false, nil
	})
	if err != nil {
		return nil, nil
	}
	return
}

func (k Keeper) IterateAllInsurances(ctx sdk.Context, cb func(insurance types.Insurance) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixInsurance)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var insurance types.Insurance
		k.cdc.MustUnmarshal(iterator.Value(), &insurance)

		stop, err := cb(insurance)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

func (k Keeper) IteratePairingInsurances(ctx sdk.Context, cb func(insurance types.Insurance) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixPairingInsuranceIndex)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		id := types.ParsePairingInsuranceIndexKey(iterator.Key())
		insurance, _ := k.GetInsurance(ctx, id)
		stop, err := cb(insurance)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

func (k Keeper) GetInsurancesFromProviderAddress(ctx sdk.Context, providerAddress sdk.AccAddress) []types.Insurance {
	var insurances []types.Insurance

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, append(types.KeyPrefixInsurancesByProviderIndex, address.MustLengthPrefix(providerAddress)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		_, insuranceId := types.ParseInsurancesByProviderIndexKey(iterator.Key())
		insurance, _ := k.GetInsurance(ctx, insuranceId)
		insurances = append(insurances, insurance)
	}

	return insurances
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
