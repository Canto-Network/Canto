package keeper

import (
	"fmt"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetNetAmountStateEssentials(ctx sdk.Context) (
	nase types.NetAmountStateEssentials, pairedChunkWithInsuranceId map[uint64]types.Chunk,
	pairedInsurances []types.Insurance, validatorMap map[string]stakingtypes.Validator,
) {
	liquidBondDenom := k.GetLiquidBondDenom(ctx)
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	totalDelShares := sdk.ZeroDec()
	totalChunksBalance := sdk.ZeroInt()
	totalRemainingRewards := sdk.ZeroDec()
	totalRemainingRewardsBeforeModuleFee := sdk.ZeroDec()
	totalLiquidTokens := sdk.ZeroInt()
	totalUnbondingChunksBalance := sdk.ZeroInt()
	numPairedChunks := sdk.ZeroInt()

	pairedChunkWithInsuranceId = make(map[uint64]types.Chunk)
	// To reduce gas consumption, store validator info in map
	validatorMap = make(map[string]stakingtypes.Validator)
	k.IterateAllChunks(ctx, func(chunk types.Chunk) (stop bool) {
		balance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
		totalChunksBalance = totalChunksBalance.Add(balance.Amount)

		switch chunk.Status {
		case types.CHUNK_STATUS_PAIRED:
			numPairedChunks = numPairedChunks.Add(sdk.OneInt())
			pairedIns := k.mustGetInsurance(ctx, chunk.PairedInsuranceId)
			pairedChunkWithInsuranceId[chunk.PairedInsuranceId] = chunk
			pairedInsurances = append(pairedInsurances, pairedIns)
			// Use map to reduce gas consumption
			if _, ok := validatorMap[pairedIns.ValidatorAddress]; !ok {
				validator, found := k.stakingKeeper.GetValidator(ctx, pairedIns.GetValidator())
				if !found {
					panic(fmt.Sprintf("validator of paired ins %s not found(insuranceId: %d)", pairedIns.GetValidator(), pairedIns.Id))
				}
				validatorMap[pairedIns.ValidatorAddress] = validator
			}
			validator := validatorMap[pairedIns.ValidatorAddress]

			// Get delegation of chunk
			del, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), validator.GetOperator())
			if !found {
				panic(fmt.Sprintf("delegation not found: %s(chunkId: %d)", chunk.DerivedAddress(), chunk.Id))
			}

			totalDelShares = totalDelShares.Add(del.GetShares())
			tokenValue := validator.TokensFromSharesTruncated(del.GetShares()).TruncateInt()
			// TODO: Currently we don't consider unpairing insurance's balance for re-pairing or re-delegation scenarios.
			tokenValue = k.calcTokenValueWithInsuranceCoverage(ctx, tokenValue, pairedIns)
			totalLiquidTokens = totalLiquidTokens.Add(tokenValue)

			beforeCachedCtxConsumed := ctx.GasMeter().GasConsumed()
			cachedCtx, _ := ctx.CacheContext()
			endingPeriod := k.distributionKeeper.IncrementValidatorPeriod(cachedCtx, validator)
			delRewards := k.distributionKeeper.CalculateDelegationRewards(cachedCtx, validator, del, endingPeriod)
			afterCachedCtxConsumed := cachedCtx.GasMeter().GasConsumed()
			cachedCtx.GasMeter().RefundGas(
				afterCachedCtxConsumed-beforeCachedCtxConsumed,
				"cachedCtx does not write state",
			)
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
	nase = types.NetAmountStateEssentials{
		LsTokensTotalSupply:         k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount,
		TotalLiquidTokens:           totalLiquidTokens,
		TotalChunksBalance:          totalChunksBalance,
		TotalDelShares:              totalDelShares,
		TotalRemainingRewards:       totalRemainingRewards,
		TotalUnbondingChunksBalance: totalUnbondingChunksBalance,
		NumPairedChunks:             numPairedChunks,
		ChunkSize:                   types.ChunkSize,
		FeeRate:                     moduleFeeRate,
		UtilizationRatio:            u,
		RewardModuleAccBalance:      rewardPoolBalance,
		RemainingChunkSlots:         types.GetAvailableChunkSlots(u, params.DynamicFeeRate.UHardCap, totalSupplyAmt),
	}
	nase.NetAmount = nase.CalcNetAmount()
	nase.MintRate = nase.CalcMintRate()
	nase.DiscountRate = nase.CalcDiscountRate(params.MaximumDiscountRate)

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
