package v7

import (
	"fmt"

	inflationkeeper "github.com/Canto-Network/Canto/v1/x/inflation/keeper"
	inflationtypes "github.com/Canto-Network/Canto/v1/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ik inflationkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	fmt.Println("CREATE UPGRADE HANDLER")
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("CANTO UPGRADE STARTED")

		fmt.Println("UPGRADE HERE")

		// UpdateInflationParams(ctx, ik)

		vm[inflationtypes.ModuleName] = 1

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func UpdateInflationParams(ctx sdk.Context, ik inflationkeeper.Keeper) {
	fmt.Println("Setting Params")
	params := ik.GetParams(ctx)
	newExp := inflationtypes.ExponentialCalculation{
		A:             sdk.NewDec(int64(16_304_348)),
		R:             sdk.NewDecWithPrec(35, 2), // 35%
		C:             sdk.ZeroDec(),
		BondingTarget: sdk.NewDecWithPrec(80, 2), // not relevant; max variance is 0
		MaxVariance:   sdk.ZeroDec(),             // 0%
	}
	params.ExponentialCalculation = newExp
}
