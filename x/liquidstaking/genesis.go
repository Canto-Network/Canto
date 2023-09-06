package liquidstaking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/keeper"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}
	stakingUnbondingTime := k.GetUnbondingTime(ctx)
	if genState.Epoch.Duration != stakingUnbondingTime {
		panic(types.ErrInvalidEpochDuration)
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
	for _, info := range genState.RedelegationInfos {
		k.SetRedelegationInfo(ctx, info)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return types.NewGenesisState(
		k.GetLiquidBondDenom(ctx),
		k.GetParams(ctx),
		k.GetEpoch(ctx),
		k.GetLastChunkId(ctx),
		k.GetLastInsuranceId(ctx),
		k.GetAllChunks(ctx),
		k.GetAllInsurances(ctx),
		k.GetAllUnpairingForUnstakingChunkInfos(ctx),
		k.GetAllWithdrawInsuranceRequests(ctx),
		k.GetAllRedelegationInfos(ctx))
}
