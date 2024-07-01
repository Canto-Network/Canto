package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

// InitGenesis initializes the coinswap module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := types.ValidateGenesis(genState); err != nil {
		panic(fmt.Errorf("panic for ValidateGenesis,%v", err))
	}

	//init to prevent nil slice for MaxSwapAmount in params
	if genState.Params.MaxSwapAmount == nil || len(genState.Params.MaxSwapAmount) == 0 {
		genState.Params.MaxSwapAmount = sdk.Coins{}
	}
	k.SetParams(ctx, genState.Params)
	k.SetStandardDenom(ctx, genState.StandardDenom)
	k.setSequence(ctx, genState.Sequence)
	for _, pool := range genState.Pool {
		k.setPool(ctx, &pool)
	}
}

// ExportGenesis returns the coinswap module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params := k.GetParams(ctx)
	//init to prevent nil slice for MaxSwapAmount in params
	if params.MaxSwapAmount == nil || len(params.MaxSwapAmount) == 0 {
		params.MaxSwapAmount = sdk.Coins{}
	}
	standardDenom, _ := k.GetStandardDenom(ctx)
	return types.GenesisState{
		Params:        params,
		StandardDenom: standardDenom,
		Pool:          k.GetAllPools(ctx),
		Sequence:      k.getSequence(ctx),
	}
}
