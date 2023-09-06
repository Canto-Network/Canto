package types

import (
	"fmt"
	"sort"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	// MinimumCollateral is the minimum collateral rate for insurance to be paired.
	// Insurance provider must provide collateral at least 7% of the chunk size tokens.
	// TODO: limit changing slashing module parameter (it shouldn't be easy to change, maybe in AnteHandler which catch param change message and make it fail?)
	MinimumCollateral = "0.07" // 7%
)

func NewInsurance(id uint64, providerAddress, validatorAddress string, feeRate sdk.Dec) Insurance {
	return Insurance{
		Id:               id,
		ChunkId:          0, // Not yet assigned
		Status:           INSURANCE_STATUS_PAIRING,
		ProviderAddress:  providerAddress,
		ValidatorAddress: validatorAddress,
		FeeRate:          feeRate,
	}
}

func (i *Insurance) DerivedAddress() sdk.AccAddress {
	return DeriveAddress(ModuleName, fmt.Sprintf("insurance%d", i.Id))
}

func (i *Insurance) FeePoolAddress() sdk.AccAddress {
	return DeriveAddress(ModuleName, fmt.Sprintf("insurancefee%d", i.Id))
}

func (i *Insurance) GetProvider() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(i.ProviderAddress)
}

func (i *Insurance) GetValidator() sdk.ValAddress {
	valAddr, _ := sdk.ValAddressFromBech32(i.ValidatorAddress)
	return valAddr
}

// SortInsurances sorts insurances by fee rate and id
// If descending is true, it sorts in descending order which means the highest fee rate comes first.
// TODO: Need memory profiling
// This can be called multiple times and there are local assignments for i, j Insurance
// readable but worried for memory usage
func SortInsurances(validatorMap map[string]stakingtypes.Validator, insurances []Insurance, descending bool) {
	sort.Slice(insurances, func(i, j int) bool {
		iInsurance := insurances[i]
		jInsurance := insurances[j]

		iValidator := validatorMap[iInsurance.ValidatorAddress]
		jValidator := validatorMap[jInsurance.ValidatorAddress]

		iFee := iValidator.Commission.Rate.Add(iInsurance.FeeRate)
		jFee := jValidator.Commission.Rate.Add(jInsurance.FeeRate)

		if !iFee.Equal(jFee) {
			if descending {
				return iFee.GT(jFee)
			}
			return iFee.LT(jFee)
		}
		if descending {
			return iInsurance.Id > jInsurance.Id
		}
		return iInsurance.Id < jInsurance.Id
	})
}

func (i *Insurance) Equal(other Insurance) bool {
	return i.Id == other.Id &&
		i.ChunkId == other.ChunkId &&
		i.Status == other.Status &&
		i.ProviderAddress == other.ProviderAddress &&
		i.ValidatorAddress == other.ValidatorAddress &&
		i.FeeRate.Equal(other.FeeRate)
}

func (i *Insurance) SetStatus(status InsuranceStatus) {
	i.Status = status
}

func (i *Insurance) Validate(lastInsuranceId uint64) error {
	if i.Id > lastInsuranceId {
		return sdkerrors.Wrapf(ErrInvalidInsuranceId, "insurance id must be %d or less", lastInsuranceId)
	}
	if i.Status == INSURANCE_STATUS_UNSPECIFIED {
		return ErrInvalidInsuranceStatus
	}
	return nil
}

func (i *Insurance) EmptyChunk() {
	i.ChunkId = Empty
}

func (i *Insurance) HasChunk() bool {
	return i.ChunkId != Empty
}

func (i *Insurance) IsUnpaired() bool {
	return i.Status == INSURANCE_STATUS_PAIRING || i.Status == INSURANCE_STATUS_UNPAIRED
}
