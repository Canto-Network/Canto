package govshuttle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Canto-Network/Canto/v8/x/govshuttle/keeper"
	"github.com/Canto-Network/Canto/v8/x/govshuttle/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, accountKeeper authkeeper.AccountKeeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	if genState.PortContractAddr != "" {
		portAddr := common.HexToAddress(genState.PortContractAddr)
		k.SetPort(ctx, portAddr)
	}

	if acc := accountKeeper.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the govshuttle module account has not been set")
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	if portAddr, ok := k.GetPort(ctx); ok {
		genesis.PortContractAddr = portAddr.String()
	}

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
