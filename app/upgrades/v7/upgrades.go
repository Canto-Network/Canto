package v7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	inflationtypes "github.com/Canto-Network/Canto/v1/x/inflation/types" 
	inflationkeeper "github.com/Canto-Network/Canto/v1/x/inflation/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ik inflationkeeper.Keeper,
)	upgradetypes.UpgradeHandler {
		return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			
			vm[inflationtypes.ModuleName] = 2

			return mm.RunMigrations(ctx, configurator, vm)
		}
}
