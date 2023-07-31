package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// CalcUtilizationRatio returns a utilization ratio of liquidstaking module.
func CalcUtilizationRatio(netAmountBeforeModuleFee sdk.Dec, totalSupplyAmt sdk.Int) sdk.Dec {
	if totalSupplyAmt.IsZero() || netAmountBeforeModuleFee.IsZero() {
		return sdk.ZeroDec()
	}
	// netAmountBeforeModuleFee / totalSupply
	return netAmountBeforeModuleFee.Quo(totalSupplyAmt.ToDec())
}

// CalcDynamicFeeRate returns a dynamic fee rate of a module
// and utilization ratio when it used to calculate the fee rate.
func CalcDynamicFeeRate(utilizationRatio sdk.Dec, dynamicFeeRate DynamicFeeRate) (
	feeRate sdk.Dec,
) {
	// set every field of params as separate variable
	r0, softCap, optimal, hardCap, slope1, slope2 := dynamicFeeRate.R0,
		dynamicFeeRate.USoftCap, dynamicFeeRate.UOptimal, dynamicFeeRate.UHardCap,
		dynamicFeeRate.Slope1, dynamicFeeRate.Slope2

	hardCap = sdk.MinDec(hardCap, SecurityCap)
	if utilizationRatio.LT(softCap) {
		feeRate = r0
		return feeRate
	}
	if utilizationRatio.LTE(optimal) {
		feeRate = CalcFormulaBetweenSoftCapAndOptimal(r0, softCap, optimal, slope1, utilizationRatio)
		return feeRate
	}
	feeRate = CalcFormulaUpperOptimal(r0, optimal, hardCap, slope1, slope2, utilizationRatio)
	return feeRate
}

// CalcFormulaBetweenSoftCapAndOptimal returns a dynamic fee rate with formula between softcap and optimal.
func CalcFormulaBetweenSoftCapAndOptimal(
	r0, softCap, optimal, slope1, u sdk.Dec,
) sdk.Dec {
	// r0 + ((u - softcap) / (optimal - softcap) x slope1)
	return r0.Add(
		u.Sub(softCap).Quo(
			optimal.Sub(softCap),
		).Mul(slope1),
	)
}

func CalcFormulaUpperOptimal(
	r0, optimal, hardCap, slope1, slope2, u sdk.Dec,
) sdk.Dec {
	// r0 + slope1 + ((min(u, hardcap) - optimal) / (hardcap - optimal) x slope2)
	return r0.Add(slope1).Add(
		sdk.MinDec(u, hardCap).Sub(optimal).Quo(
			hardCap.Sub(optimal),
		).Mul(slope2))
}

// GetAvailableChunkSlots returns a number of chunk which can be paired.
func GetAvailableChunkSlots(u, uHardCap sdk.Dec, totalSupplyAmt sdk.Int) sdk.Int {
	hardCap := sdk.MinDec(uHardCap, SecurityCap)
	remainingU := hardCap.Sub(u)
	if !remainingU.IsPositive() {
		return sdk.ZeroInt()
	}
	return remainingU.Mul(totalSupplyAmt.ToDec()).QuoTruncate(ChunkSize.ToDec()).TruncateInt()
}
