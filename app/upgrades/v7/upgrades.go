package v7

import (
	coinswapkeeper "github.com/Canto-Network/Canto/v7/x/coinswap/keeper"
	onboardingkeeper "github.com/Canto-Network/Canto/v7/x/onboarding/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	onboardingKeeper onboardingkeeper.Keeper,
	coinswapKeeper coinswapkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrading to v7.0.0", UpgradeName)

		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		params := onboardingKeeper.GetParams(ctx)
		params.WhitelistedChannels = []string{"channel-0"}
		params.AutoSwapThreshold = sdk.NewIntWithDecimal(4, 18)
		onboardingKeeper.SetParams(ctx, params)

		coinswapParams := coinswapKeeper.GetParams(ctx)
		coinswapParams.PoolCreationFee = sdk.NewCoin("acanto", sdk.ZeroInt())
		coinswapParams.MaxSwapAmount = sdk.NewCoins(
			sdk.NewCoin(UsdcIBCDenom, sdk.NewIntWithDecimal(10, 6)),
			sdk.NewCoin(UsdtIBCDenom, sdk.NewIntWithDecimal(10, 6)),
			sdk.NewCoin(EthIBCDenom, sdk.NewIntWithDecimal(1, 17)),
		)
		coinswapKeeper.SetParams(ctx, coinswapParams)
		coinswapKeeper.SetStandardDenom(ctx, "acanto")

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return newVM, nil
	}
}
