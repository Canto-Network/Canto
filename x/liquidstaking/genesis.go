package liquidstaking

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetLiquidBondDenom(ctx, genState.LiquidBondDenom)
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var chunks []types.Chunk
	err := k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		chunks = append(chunks, chunk)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	var insurances []types.Insurance
	err = k.IterateAllInsurances(ctx, func(insurance types.Insurance) (bool, error) {
		insurances = append(insurances, insurance)
		return false, nil
	})

	genesis := types.DefaultGenesisState()
	genesis.LiquidBondDenom = k.GetLiquidBondDenom(ctx)
	genesis.Params = k.GetParams(ctx)
	genesis.Epoch = k.GetEpoch(ctx)
	genesis.LastChunkId = k.GetLastChunkId(ctx)
	genesis.LastInsuranceId = k.GetLastInsuranceId(ctx)
	genesis.Chunks = chunks
	genesis.Insurances = insurances
	genesis.PendingLiquidUnstakes = k.GetAllPendingLiquidUnstake(ctx)
	genesis.UnpairingForUnstakeChunkInfos = k.GetAllUnpairingForUnstakeChunkInfos(ctx)
	genesis.WithdrawInsuranceRequests = k.GetAllWithdrawInsuranceRequests(ctx)

	return genesis
}
