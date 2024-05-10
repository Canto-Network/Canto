package v8

import (
	"context"

	"cosmossdk.io/collections"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v8
func CreateUpgradeHandler(
	mm *module.Manager,
	legacySubspace paramstypes.Subspace,
	consensusParamsStore collections.Item[types.ConsensusParams],
	configurator module.Configurator,
	clientKeeper clientkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrade: ", UpgradeName)

		// ibc-go vX -> v6
		// - skip
		// - not implement an ICS27 controller module
		//
		// ibc-go v6 -> v7
		// - skip
		// - pruning expired tendermint consensus states is optional
		//
		// ibc-go v7 -> v7.1
		// - apply
		{
			// explicitly update the IBC 02-client params, adding the localhost client type
			params := clientKeeper.GetParams(sdkCtx)
			params.AllowedClients = append(params.AllowedClients, exported.Localhost)
			clientKeeper.SetParams(sdkCtx, params)
		}

		if err := baseapp.MigrateParams(sdkCtx, legacySubspace, consensusParamsStore); err != nil {
			return vm, err
		}
		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
