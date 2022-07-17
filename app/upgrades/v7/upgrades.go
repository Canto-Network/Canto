package v7

import (
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
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Canto v2 upgrade started")

		vm[inflationtypes.ModuleName] = 1

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
