package v7

import (
	"github.com/Canto-Network/Canto/v6/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type InflationKeeper interface {
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)
	SetEpochsPerPeriod(ctx sdk.Context, epochsPerPeriod int64)
	SetEpochMintProvision(ctx sdk.Context, epochMintProvision sdk.Dec)
}

func UpdateParams(ctx sdk.Context, ik InflationKeeper) error {
	params := ik.GetParams(ctx)
	newExp := types.ExponentialCalculation{
		A:             sdk.NewDec(int64(16_304_348)),
		R:             sdk.NewDecWithPrec(35, 2), // 35%
		C:             sdk.ZeroDec(),
		BondingTarget: sdk.NewDecWithPrec(80, 2), // not relevant; max variance is 0
		MaxVariance:   sdk.ZeroDec(),
	}

	ctx.Logger().Info("Setting Inflation Parameters")

	params.ExponentialCalculation = newExp
	ik.SetParams(ctx, params)

	//update EpochsPerPeriod
	ik.SetEpochsPerPeriod(ctx, int64(30))

	epochMintProvision := types.CalculateEpochMintProvision(params, 0, 30, sdk.NewDec(1))
	ik.SetEpochMintProvision(ctx, epochMintProvision)

	return nil
}
