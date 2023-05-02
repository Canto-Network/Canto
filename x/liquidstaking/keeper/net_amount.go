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
	totalLiquidTokens := sdk.ZeroInt()
	totalInsuranceTokens := sdk.ZeroInt()
	totalUnbondingBalance := sdk.ZeroDec()

	err := k.IterateAllChunks(ctx, func(chunk types.Chunk) (stop bool, err error) {
		balance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
		totalChunksBalance = totalChunksBalance.Add(balance.Amount.ToDec())

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
		totalRemainingRewards = totalRemainingRewards.Add(delReward.AmountOf(bondDenom))

		k.stakingKeeper.IterateDelegatorUnbondingDelegations(ctx, chunk.DerivedAddress(), func(ubd stakingtypes.UnbondingDelegation) (stop bool) {
			for _, entry := range ubd.Entries {
				totalUnbondingBalance = totalUnbondingBalance.Add(entry.Balance.ToDec())
			}
			return false
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	// Iterate all paired insurances to get total insurance tokens
	err = k.IterateAllInsurances(ctx, func(insurance types.Insurance) (stop bool, err error) {
		if insurance.Status == types.INSURANCE_STATUS_PAIRED {
			insuranceBalance := k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
			totalInsuranceTokens = totalInsuranceTokens.Add(insuranceBalance.Amount)
		}
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	nas = types.NetAmountState{
		LsTokensTotalSupply:    k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount,
		TotalChunksBalance:     totalChunksBalance.TruncateInt(),
		TotalDelShares:         totalDelShares,
		TotalRemainingRewards:  totalRemainingRewards,
		TotalLiquidTokens:      totalLiquidTokens,
		TotalInsuranceTokens:   totalInsuranceTokens,
		TotalUnbondingBalance:  totalUnbondingBalance.TruncateInt(),
		RewardModuleAccBalance: k.bankKeeper.GetBalance(ctx, types.RewardPool, bondDenom).Amount,
	}

	nas.NetAmount = nas.CalcNetAmount(k.bankKeeper.GetBalance(ctx, types.RewardPool, bondDenom).Amount)
	nas.MintRate = nas.CalcMintRate()
	return
}
