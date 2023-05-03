package v6

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	coinswapkeeper "github.com/Canto-Network/Canto/v6/x/coinswap/keeper"
	coinswaptypes "github.com/Canto-Network/Canto/v6/x/coinswap/types"
	onboardingkeeper "github.com/Canto-Network/Canto/v6/x/onboarding/keeper"
	onboardingtypes "github.com/Canto-Network/Canto/v6/x/onboarding/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	onboardingKeeper onboardingkeeper.Keeper,
	coinswapKeeper coinswapkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrading to v6.0.0", UpgradeName)

		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		onboardingParams := onboardingtypes.DefaultParams()
		onboardingKeeper.SetParams(ctx, onboardingParams)

		coinswapParams := coinswaptypes.DefaultParams()
		coinswapKeeper.SetParams(ctx, coinswapParams)
		coinswapKeeper.SetStandardDenom(ctx, "acanto")

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return newVM, nil
	}
}
