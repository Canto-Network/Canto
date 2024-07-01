package v7

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	coinswapkeeper "github.com/Canto-Network/Canto/v7/x/coinswap/keeper"
	coinswaptypes "github.com/Canto-Network/Canto/v7/x/coinswap/types"
	onboardingkeeper "github.com/Canto-Network/Canto/v7/x/onboarding/keeper"
	onboardingtypes "github.com/Canto-Network/Canto/v7/x/onboarding/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v7
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	onboardingKeeper onboardingkeeper.Keeper,
	coinswapKeeper coinswapkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrading to v7.0.0", UpgradeName)

		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		onboardingParams := onboardingtypes.DefaultParams()
		onboardingKeeper.SetParams(sdkCtx, onboardingParams)

		coinswapParams := coinswaptypes.DefaultParams()
		coinswapKeeper.SetParams(sdkCtx, coinswapParams)
		coinswapKeeper.SetStandardDenom(sdkCtx, "acanto")

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return newVM, nil
	}
}
