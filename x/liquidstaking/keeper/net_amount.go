package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO: Discuss with taeyoung what values should be used for meaningful testing
func (k Keeper) GetNetAmountState(ctx sdk.Context) (nas types.NetAmountState) {
	liquidBondDenom := k.GetLiquidBondDenom(ctx)
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	totalDelShares := sdk.ZeroDec()
	totalChunksBalance := sdk.NewDec(0)
	totalRemainingRewards := sdk.ZeroDec()
	totalRemainingInsuranceCommissions := sdk.ZeroDec()
	totalLiquidTokens := sdk.ZeroInt()
	totalInsuranceTokens := sdk.ZeroInt()
	totalInsuranceCommissions := sdk.ZeroInt()
	totalPairedInsuranceTokens := sdk.ZeroInt()
	totalPairedInsuranceCommissions := sdk.ZeroInt()
	totalUnpairingInsuranceTokens := sdk.ZeroInt()
	totalUnpairingInsuranceCommissions := sdk.ZeroInt()
	totalUnpairedInsuranceTokens := sdk.ZeroInt()
	totalUnpairedInsuranceCommissions := sdk.ZeroInt()
	totalUnbondingBalance := sdk.ZeroDec()

	err := k.IterateAllChunks(ctx, func(chunk types.Chunk) (stop bool, err error) {
		balance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
		totalChunksBalance = totalChunksBalance.Add(balance.Amount.ToDec())

		if chunk.PairedInsuranceId != 0 {
			// chunk is paired which meanas have delegation
			pairedInsurance, _ := k.GetInsurance(ctx, chunk.PairedInsuranceId)
			valAddr, err := sdk.ValAddressFromBech32(pairedInsurance.ValidatorAddress)
			if err != nil {
				return true, err
			}
			validator := k.stakingKeeper.Validator(ctx, valAddr)
			delegation, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), valAddr)
			if !found {
				return false, nil
			}
			totalDelShares = totalDelShares.Add(delegation.GetShares())
			tokens := validator.TokensFromSharesTruncated(delegation.GetShares()).TruncateInt()
			totalLiquidTokens = totalLiquidTokens.Add(tokens)
			cachedCtx, _ := ctx.CacheContext()
			endingPeriod := k.distributionKeeper.IncrementValidatorPeriod(cachedCtx, validator)
			delReward := k.distributionKeeper.CalculateDelegationRewards(cachedCtx, validator, delegation, endingPeriod)
			insuranceCommission := delReward.MulDec(pairedInsurance.FeeRate)
			totalRemainingInsuranceCommissions = totalRemainingInsuranceCommissions.Add(insuranceCommission.AmountOf(bondDenom))
			// insuranceCommission is not reward of module
			pureReward := delReward.Sub(insuranceCommission)
			totalRemainingRewards = totalRemainingRewards.Add(pureReward.AmountOf(bondDenom))
		} else {
			k.stakingKeeper.IterateDelegatorUnbondingDelegations(ctx, chunk.DerivedAddress(), func(ubd stakingtypes.UnbondingDelegation) (stop bool) {
				for _, entry := range ubd.Entries {
					totalUnbondingBalance = totalUnbondingBalance.Add(entry.Balance.ToDec())
				}
				return false
			})
		}
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	// Iterate all paired insurances to get total insurance tokens
	err = k.IterateAllInsurances(ctx, func(insurance types.Insurance) (stop bool, err error) {
		insuranceBalance := k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), bondDenom)
		commission := k.bankKeeper.GetBalance(ctx, insurance.FeePoolAddress(), bondDenom)
		switch insurance.Status {
		case types.INSURANCE_STATUS_PAIRED:
			totalPairedInsuranceTokens = totalPairedInsuranceTokens.Add(insuranceBalance.Amount)
			totalPairedInsuranceCommissions = totalPairedInsuranceCommissions.Add(commission.Amount)
		case types.INSURANCE_STATUS_UNPAIRING:
			totalUnpairingInsuranceTokens = totalUnpairingInsuranceTokens.Add(insuranceBalance.Amount)
			totalUnpairingInsuranceCommissions = totalUnpairingInsuranceCommissions.Add(commission.Amount)
		case types.INSURANCE_STATUS_UNPAIRED:
			totalUnpairedInsuranceTokens = totalUnpairedInsuranceTokens.Add(insuranceBalance.Amount)
			totalUnpairedInsuranceCommissions = totalUnpairedInsuranceCommissions.Add(commission.Amount)
		}
		totalInsuranceTokens = totalInsuranceTokens.Add(insuranceBalance.Amount)
		totalInsuranceCommissions = totalInsuranceCommissions.Add(commission.Amount)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	nas = types.NetAmountState{
		LsTokensTotalSupply:                k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount,
		TotalChunksBalance:                 totalChunksBalance.TruncateInt(),
		TotalDelShares:                     totalDelShares,
		TotalRemainingRewards:              totalRemainingRewards,
		TotalRemainingInsuranceCommissions: totalRemainingInsuranceCommissions,
		TotalLiquidTokens:                  totalLiquidTokens,
		TotalInsuranceTokens:               totalInsuranceTokens,
		TotalInsuranceCommissions:          totalInsuranceCommissions,
		TotalPairedInsuranceTokens:         totalPairedInsuranceTokens,
		TotalPairedInsuranceCommissions:    totalPairedInsuranceCommissions,
		TotalUnpairingInsuranceTokens:      totalUnpairingInsuranceTokens,
		TotalUnpairingInsuranceCommissions: totalUnpairingInsuranceCommissions,
		TotalUnpairedInsuranceTokens:       totalUnpairedInsuranceTokens,
		TotalUnpairedInsuranceCommissions:  totalUnpairedInsuranceCommissions,
		TotalUnbondingBalance:              totalUnbondingBalance.TruncateInt(),
		RewardModuleAccBalance:             k.bankKeeper.GetBalance(ctx, types.RewardPool, bondDenom).Amount,
	}

	nas.NetAmount = nas.CalcNetAmount(k.bankKeeper.GetBalance(ctx, types.RewardPool, bondDenom).Amount)
	nas.MintRate = nas.CalcMintRate()
	return
}
