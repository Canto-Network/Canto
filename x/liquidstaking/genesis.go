package liquidstaking

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}
	k.SetParams(ctx, genState.Params)
	k.SetEpoch(ctx, genState.Epoch)
	k.SetLiquidBondDenom(ctx, genState.LiquidBondDenom)
	k.SetLastChunkId(ctx, genState.LastChunkId)
	k.SetLastInsuranceId(ctx, genState.LastInsuranceId)
	for _, chunk := range genState.Chunks {
		k.SetChunk(ctx, chunk)
	}
	for _, insurance := range genState.Insurances {
		k.SetInsurance(ctx, insurance)
	}
	for _, UnpairingForUnstakingChunkInfo := range genState.UnpairingForUnstakingChunkInfos {
		k.SetUnpairingForUnstakingChunkInfo(ctx, UnpairingForUnstakingChunkInfo)
	}
	for _, request := range genState.WithdrawInsuranceRequests {
		k.SetWithdrawInsuranceRequest(ctx, request)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesisState()
	genesis.LiquidBondDenom = k.GetLiquidBondDenom(ctx)
	genesis.Params = k.GetParams(ctx)
	genesis.Epoch = k.GetEpoch(ctx)
	genesis.LastChunkId = k.GetLastChunkId(ctx)
	genesis.LastInsuranceId = k.GetLastInsuranceId(ctx)
	genesis.Chunks = k.GetAllChunks(ctx)
	genesis.Insurances = k.GetAllInsurances(ctx)
	genesis.UnpairingForUnstakingChunkInfos = k.GetAllUnpairingForUnstakingChunkInfos(ctx)
	genesis.WithdrawInsuranceRequests = k.GetAllWithdrawInsuranceRequests(ctx)

	return genesis
}
