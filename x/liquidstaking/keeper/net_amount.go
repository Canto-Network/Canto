package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetNetAmountState(ctx sdk.Context) (nas types.NetAmountState) {
	liquidBondDenom := k.GetLiquidBondDenom(ctx)
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	totalDelShares := sdk.ZeroDec()
	totalChunksBalance := sdk.ZeroInt()
	totalRemainingRewards := sdk.ZeroDec()
	totalRemainingRewardsBeforeModuleFee := sdk.ZeroDec()
	totalRemainingInsuranceCommissions := sdk.ZeroDec()
	totalLiquidTokens := sdk.ZeroInt()
	totalInsuranceTokens := sdk.ZeroInt()
	totalPairedInsuranceTokens := sdk.ZeroInt()
	totalUnpairingInsuranceTokens := sdk.ZeroInt()
	totalUnbondingChunksBalance := sdk.ZeroInt()
	numPairedChunks := sdk.ZeroInt()

	k.IterateAllInsurances(ctx, func(insurance types.Insurance) (stop bool) {
		insuranceBalance := k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), bondDenom)
		commission := k.bankKeeper.GetBalance(ctx, insurance.FeePoolAddress(), bondDenom)
		switch insurance.Status {
		case types.INSURANCE_STATUS_PAIRED:
			totalPairedInsuranceTokens = totalPairedInsuranceTokens.Add(insuranceBalance.Amount)
		case types.INSURANCE_STATUS_UNPAIRING:
			totalUnpairingInsuranceTokens = totalUnpairingInsuranceTokens.Add(insuranceBalance.Amount)
		case types.INSURANCE_STATUS_UNPAIRED:
		}
		totalInsuranceTokens = totalInsuranceTokens.Add(insuranceBalance.Amount)
		totalRemainingInsuranceCommissions = totalRemainingInsuranceCommissions.Add(commission.Amount.ToDec())
		return false
	})
	k.IterateAllChunks(ctx, func(chunk types.Chunk) (stop bool) {
		balance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
		totalChunksBalance = totalChunksBalance.Add(balance.Amount)

		switch chunk.Status {
		case types.CHUNK_STATUS_PAIRED:
			numPairedChunks = numPairedChunks.Add(sdk.OneInt())
			pairedIns, validator, del := k.mustValidatePairedChunk(ctx, chunk)
			totalDelShares = totalDelShares.Add(del.GetShares())
			tokenValue := validator.TokensFromSharesTruncated(del.GetShares()).TruncateInt()
			// TODO: Currently we don't consider unpairing insurance's balance for re-pairing or re-delegation scenarios.
			tokenValue = k.calcTokenValueWithInsuranceCoverage(ctx, tokenValue, pairedIns)
			totalLiquidTokens = totalLiquidTokens.Add(tokenValue)

			cachedCtx, _ := ctx.CacheContext()
			endingPeriod := k.distributionKeeper.IncrementValidatorPeriod(cachedCtx, validator)
			delRewards := k.distributionKeeper.CalculateDelegationRewards(cachedCtx, validator, del, endingPeriod)
			// chunk's remaining reward is calculated by
			// 1. rest = del_reward - insurance_commission
			// 2. remaining = rest x (1 - module_fee_rate)
			delReward := delRewards.AmountOf(bondDenom)
			insuranceCommission := delReward.Mul(pairedIns.FeeRate)
			remainingReward := delReward.Sub(insuranceCommission)
			totalRemainingRewardsBeforeModuleFee = totalRemainingRewardsBeforeModuleFee.Add(remainingReward)
		default:
			k.stakingKeeper.IterateDelegatorUnbondingDelegations(ctx, chunk.DerivedAddress(), func(ubd stakingtypes.UnbondingDelegation) (stop bool) {
				for _, entry := range ubd.Entries {
					unpairingIns := k.mustGetInsurance(ctx, chunk.UnpairingInsuranceId)
					tokenValue := k.calcTokenValueWithInsuranceCoverage(ctx, entry.Balance, unpairingIns)
					totalUnbondingChunksBalance = totalUnbondingChunksBalance.Add(tokenValue)
				}
				return false
			})
		}
		return false
	})
	rewardPoolBalance := k.bankKeeper.GetBalance(ctx, types.RewardPool, bondDenom).Amount
	netAmountBeforeModuleFee := rewardPoolBalance.Add(totalChunksBalance).
		Add(totalLiquidTokens).
		Add(totalUnbondingChunksBalance).ToDec().
		Add(totalRemainingRewardsBeforeModuleFee)
	totalSupplyAmt := k.bankKeeper.GetSupply(ctx, bondDenom).Amount
	params := k.GetParams(ctx)
	u := types.CalcUtilizationRatio(netAmountBeforeModuleFee, totalSupplyAmt)
	moduleFeeRate := types.CalcDynamicFeeRate(u, params.DynamicFeeRate)
	totalRemainingRewards = totalRemainingRewardsBeforeModuleFee.Mul(sdk.OneDec().Sub(moduleFeeRate))
	nas = types.NetAmountState{
		LsTokensTotalSupply:                k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount,
		TotalLiquidTokens:                  totalLiquidTokens,
		TotalChunksBalance:                 totalChunksBalance,
		TotalDelShares:                     totalDelShares,
		TotalRemainingRewards:              totalRemainingRewards,
		TotalUnbondingChunksBalance:        totalUnbondingChunksBalance,
		NumPairedChunks:                    numPairedChunks,
		ChunkSize:                          types.ChunkSize,
		TotalRemainingInsuranceCommissions: totalRemainingInsuranceCommissions,
		TotalInsuranceTokens:               totalInsuranceTokens,
		TotalPairedInsuranceTokens:         totalPairedInsuranceTokens,
		TotalUnpairingInsuranceTokens:      totalUnpairingInsuranceTokens,
		FeeRate:                            moduleFeeRate,
		UtilizationRatio:                   u,
		RewardModuleAccBalance:             rewardPoolBalance,
		RemainingChunkSlots:                types.GetAvailableChunkSlots(u, params.DynamicFeeRate.UHardCap, totalSupplyAmt),
	}
	nas.NetAmount = nas.CalcNetAmount()
	nas.MintRate = nas.CalcMintRate()
	nas.DiscountRate = nas.CalcDiscountRate(params.MaximumDiscountRate)
	return
}

func (k Keeper) calcTokenValueWithInsuranceCoverage(
	ctx sdk.Context, tokenValue sdk.Int, ins types.Insurance) sdk.Int {
	penaltyAmt := types.ChunkSize.Sub(tokenValue)
	// If penaltyAmt > 0 and paired insurance can cover it, then token value is same with ChunkSize
	if penaltyAmt.IsPositive() {
		// Consider insurance coverage
		insBal := k.bankKeeper.GetBalance(ctx, ins.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
		penaltyAmt = penaltyAmt.Sub(insBal.Amount)
		if penaltyAmt.IsPositive() {
			// It means insurance can't cover penalty perfectly
			tokenValue = tokenValue.Sub(penaltyAmt)
		}
	}
	return tokenValue
}
