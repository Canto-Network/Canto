package v7 


import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Canto-Network/Canto/v1/x/inflation/types"
)

type InflationKeeper interface {
	SetParams(ctx sdk.Context, params types.Params)
	GetParams(ctx sdk.Context) types.Params
	SetEpochsPerPeriod(ctx sdk.Context, epochsPerPeriod int64) 
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
	
	fmt.Println("fmt")

	params.ExponentialCalculation = newExp
	ik.SetParams(ctx, params)
	//update EpochsPerPeriod
	ik.SetEpochsPerPeriod(ctx, int64(30))

	return nil
}