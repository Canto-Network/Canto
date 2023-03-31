package v6

import (
	"fmt"

	globalfeetypes "github.com/Canto-Network/Canto/v6/x/globalfee/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)


const (
	// TODO: does testnet have a different denom? If so I need to write a function based off
	// TODO: of the chain-id to determine the denom.
	nativeDenom = "acanto"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v6
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	paramsKeeper paramskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)

		// GlobalFee
		minGasPrices := sdk.DecCoins{
			// TODO: 0.0025acanto // Not sure what the current rate is. So we will want to modify this
			sdk.NewDecCoinFromDec(nativeDenom, sdk.NewDecWithPrec(25, 4)),
		}
		s, ok := paramsKeeper.GetSubspace(globalfeetypes.ModuleName)
		if !ok {
			panic("global fee params subspace not found")
		}
		s.Set(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, minGasPrices)
		logger.Info(fmt.Sprintf("upgraded global fee params to %s", minGasPrices))

		return versionMap, err
	}
}
