package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalcUtilizationRatio returns a utilization ratio of liquidstaking module.
func (k Keeper) CalcUtilizationRatio(ctx sdk.Context) sdk.Dec {
	totalSupply := k.bankKeeper.GetSupply(ctx, k.stakingKeeper.BondDenom(ctx))
	var numPairedChunks int64 = 0
	k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		if chunk.Status != types.CHUNK_STATUS_PAIRED {
			return false, nil
		}
		numPairedChunks++
		return false, nil
	})
	// chunkSize * numPairedChunks / totalSupply
	return types.ChunkSize.Mul(sdk.NewInt(numPairedChunks)).ToDec().Quo(totalSupply.Amount.ToDec())
}

func (k Keeper) CalcDynamicFeeRate(ctx sdk.Context) sdk.Dec {
	dynamicFeeParams := k.GetParams(ctx).DynamicFeeRate
	// set every field of params as separate variable
	r0, softCap, optimal, hardCap, slope1, slope2 := dynamicFeeParams.R0,
		dynamicFeeParams.USoftCap, dynamicFeeParams.UOptimal, dynamicFeeParams.UHardCap,
		dynamicFeeParams.Slope1, dynamicFeeParams.Slope2

	hardCap = sdk.MinDec(hardCap, types.SecurityCap)
	u := k.CalcUtilizationRatio(ctx)
	if u.LT(softCap) {
		return r0
	}
	if u.LTE(optimal) {
		return k.CalcFormulaBetweenSoftCapAndOptimal(r0, u, softCap, optimal, slope1)
	}
	return k.CalcFormulaUpperOptimal(r0, u, optimal, hardCap, slope1, slope2)
}

// CalcFormulaBetweenSoftCapAndOptimal returns a dynamic fee rate with formula between softcap and optimal.
func (k Keeper) CalcFormulaBetweenSoftCapAndOptimal(
	r0, softCap, optimal, slope1, u sdk.Dec,
) sdk.Dec {
	// r0 + ((u - softcap) / (optimal - softcap) x slope1)
	return r0.Add(
		u.Sub(softCap).Quo(
			optimal.Sub(softCap),
		).Mul(slope1),
	)
}

func (k Keeper) CalcFormulaUpperOptimal(
	r0, optimal, hardCap, slope1, slope2, u sdk.Dec,
) sdk.Dec {
	// r0 + slope1 + ((min(u, hardcap) - optimal) / (hardcap - optimal) x slope2)
	return r0.Add(slope1).Add(
		sdk.MinDec(u, hardCap).Sub(optimal).Quo(
			hardCap.Sub(optimal),
		).Mul(slope2))
}

// MaxPairedChunks returns a upper limit of paired chunks.
func (k Keeper) MaxPairedChunks(ctx sdk.Context) sdk.Int {
	hardCap := sdk.MinDec(k.GetParams(ctx).DynamicFeeRate.UHardCap, types.SecurityCap)
	totalSupply := k.bankKeeper.GetSupply(ctx, k.stakingKeeper.BondDenom(ctx))
	// 1. u = (chunkSize * numPairedChunks) / totalSupply
	// 2. numPairedChunks = u * (totalSupply / chunkSize)
	// 3. maxPairedChunks = hardCap * (totalSupply / chunkSize)
	return hardCap.Mul(totalSupply.Amount.ToDec().Quo(types.ChunkSize.ToDec())).TruncateInt()
}
