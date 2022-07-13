package v7 


import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Canto-Network/Canto/v1/x/inflation/types"
)

type InflationKeeper interface {
	SetParams(ctx sdk.Context, params types.Params)
	GetParams(ctx sdk.Context) types.Params
}

func UpdateParams(ctx sdk.Context, ik InflationKeeper) error {
	params := ik.GetParams(ctx)
	newExp := types.ExponentialCalculation {
		A:             sdk.NewDec(int64(16_304_348)),
		R:             sdk.NewDecWithPrec(35, 2), // 35%
		C:             sdk.ZeroDec(),
		BondingTarget: sdk.NewDecWithPrec(80, 2), // not relevant; max variance is 0
		MaxVariance:   sdk.ZeroDec(),   
	}

	params.ExponentialCalculation = newExp
	ik.SetParams(ctx, params)
	return nil
}