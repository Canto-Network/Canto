package v8

import (
	liquidstakingkeeper "github.com/Canto-Network/Canto/v7/x/liquidstaking/keeper"
	liquidstakingtypes "github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v8
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	liquidstakingKeeper liquidstakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrading to v8.0.0", UpgradeName)

		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		params := liquidstakingtypes.DefaultParams()
		liquidstakingKeeper.SetParams(ctx, params)

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return newVM, nil
	}
}
