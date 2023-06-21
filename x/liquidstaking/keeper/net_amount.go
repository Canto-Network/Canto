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
	totalChunksBalance := sdk.NewDec(0)
	totalRemainingRewards := sdk.ZeroDec()
	totalRemainingInsuranceCommissions := sdk.ZeroDec()
	totalLiquidTokens := sdk.ZeroInt()
	totalInsuranceTokens := sdk.ZeroInt()
	totalPairedInsuranceTokens := sdk.ZeroInt()
	totalUnpairingInsuranceTokens := sdk.ZeroInt()
	totalUnbondingChunksBalance := sdk.ZeroDec()
	numPairedChunks := sdk.ZeroInt()

	err := k.IterateAllChunks(ctx, func(chunk types.Chunk) (stop bool, err error) {
		balance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
		totalChunksBalance = totalChunksBalance.Add(balance.Amount.ToDec())

		switch chunk.Status {
		case types.CHUNK_STATUS_PAIRED:
			numPairedChunks = numPairedChunks.Add(sdk.OneInt())
			// chunk is paired which means have delegation
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
			delRewards := k.distributionKeeper.CalculateDelegationRewards(cachedCtx, validator, delegation, endingPeriod)
			delReward := delRewards.AmountOf(bondDenom)
			insuranceCommission := delReward.Mul(pairedInsurance.FeeRate)
			// insuranceCommission is not reward of module
			pureReward := delReward.Sub(insuranceCommission)
			totalRemainingRewards = totalRemainingRewards.Add(pureReward)
		default:
			k.stakingKeeper.IterateDelegatorUnbondingDelegations(ctx, chunk.DerivedAddress(), func(ubd stakingtypes.UnbondingDelegation) (stop bool) {
				for _, entry := range ubd.Entries {
					totalUnbondingChunksBalance = totalUnbondingChunksBalance.Add(entry.Balance.ToDec())
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
		case types.INSURANCE_STATUS_UNPAIRING:
			totalUnpairingInsuranceTokens = totalUnpairingInsuranceTokens.Add(insuranceBalance.Amount)
		case types.INSURANCE_STATUS_UNPAIRED:
		}
		totalInsuranceTokens = totalInsuranceTokens.Add(insuranceBalance.Amount)
		totalRemainingInsuranceCommissions = totalRemainingInsuranceCommissions.Add(commission.Amount.ToDec())
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	nas = types.NetAmountState{
		LsTokensTotalSupply:                k.bankKeeper.GetSupply(ctx, liquidBondDenom).Amount,
		TotalLiquidTokens:                  totalLiquidTokens,
		TotalChunksBalance:                 totalChunksBalance.TruncateInt(),
		TotalDelShares:                     totalDelShares,
		TotalRemainingRewards:              totalRemainingRewards,
		TotalUnbondingChunksBalance:        totalUnbondingChunksBalance.TruncateInt(),
		NumPairedChunks:                    numPairedChunks,
		ChunkSize:                          types.ChunkSize,
		TotalRemainingInsuranceCommissions: totalRemainingInsuranceCommissions,
		TotalInsuranceTokens:               totalInsuranceTokens,
		TotalPairedInsuranceTokens:         totalPairedInsuranceTokens,
		TotalUnpairingInsuranceTokens:      totalUnpairingInsuranceTokens,
	}
	nas.NetAmount = nas.CalcNetAmount(k.bankKeeper.GetBalance(ctx, types.RewardPool, bondDenom).Amount)
	nas.MintRate = nas.CalcMintRate()
	nas.RewardModuleAccBalance = k.bankKeeper.GetBalance(ctx, types.RewardPool, bondDenom).Amount
	nas.FeeRate, nas.UtilizationRatio = k.CalcDynamicFeeRate(ctx)
	nas.RemainingChunkSlots = k.GetAvailableChunkSlots(ctx)
	nas.DiscountRate = k.CalcDiscountRate(ctx)
	return
}
