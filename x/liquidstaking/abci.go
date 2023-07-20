package liquidstaking

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	if k.IsEpochReached(ctx) {
		k.CoverRedelegationPenalty(ctx)
	}
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if k.IsEpochReached(ctx) {
		k.DistributeReward(ctx)
		k.CoverSlashingAndHandleMatureUnbondings(ctx)
		k.RemoveDeletableRedelegationInfos(ctx)
		k.HandleQueuedLiquidUnstakes(ctx)
		k.HandleUnprocessedQueuedLiquidUnstakes(ctx)
		k.HandleQueuedWithdrawInsuranceRequests(ctx)
		newlyRankedInInsurances, rankOutInsurances := k.RankInsurances(ctx)
		k.RePairRankedInsurances(ctx, newlyRankedInInsurances, rankOutInsurances)
		k.IncrementEpoch(ctx)
	}
}
