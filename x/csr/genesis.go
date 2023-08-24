package csr

import (
	"github.com/Canto-Network/Canto/v7/x/csr/keeper"
	"github.com/Canto-Network/Canto/v7/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/ethereum/go-ethereum/common"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, accountKeeper authkeeper.AccountKeeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
	for _, csr := range genState.Csrs {
		k.SetCSR(ctx, csr)
	}
	if genState.TurnstileAddress != nil {
		k.SetTurnstile(ctx, common.BytesToAddress(genState.TurnstileAddress))
	}
	// make sure that the csr module account is set on genesis
	if acc := accountKeeper.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		// NOTE: shouldn't occur
		panic("the csr module account has not been set")
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	csrs := k.GetAllCSRs(ctx)

	if len(csrs) == 0 {
		genesis.Csrs = []types.CSR{}
	} else {
		genesis.Csrs = csrs
	}

	turnstileAddress, found := k.GetTurnstile(ctx)
	if found {
		genesis.TurnstileAddress = turnstileAddress.Bytes()
	} else {
		genesis.TurnstileAddress = nil
	}

	return genesis
}
