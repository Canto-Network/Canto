package v4

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	shuttleKeeper "github.com/Canto-Network/Canto/v7/x/govshuttle/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/ethereum/go-ethereum/common"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	shuttlekeeper shuttleKeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrading to v4.0.0", UpgradeName)

		// update address of map contract in keeper state
		shuttlekeeper.SetPort(sdkCtx, common.BytesToAddress(portAddr))
		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
