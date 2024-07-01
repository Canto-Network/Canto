package v8

import (
	"context"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
)

var MinCommissionRate = sdkmath.LegacyNewDecWithPrec(5, 2) // 5%

// CreateUpgradeHandler creates an SDK upgrade handler for v8
func CreateUpgradeHandler(
	mm *module.Manager,
	legacySubspace paramstypes.Subspace,
	consensusParamsStore collections.Item[types.ConsensusParams],
	configurator module.Configurator,
	clientKeeper clientkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrade: ", UpgradeName)

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		if vm, err := mm.RunMigrations(ctx, configurator, vm); err != nil {
			return vm, err
		}

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

		// canto v8 custom
		{
			params, err := stakingKeeper.GetParams(ctx)
			if err != nil {
				return vm, err
			}
			params.MinCommissionRate = MinCommissionRate
			stakingKeeper.SetParams(ctx, params)
		}

		return vm, nil
	}
}
