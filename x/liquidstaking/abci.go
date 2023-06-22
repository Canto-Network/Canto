package liquidstaking

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if k.IsEpochReached(ctx) {
		k.DistributeReward(ctx)
		k.CoverSlashingAndHandleMatureUnbondings(ctx)
		if _, err := k.HandleQueuedLiquidUnstakes(ctx); err != nil {
			panic(err)
		}
		if err := k.HandleUnprocessedQueuedLiquidUnstakes(ctx); err != nil {
			panic(err)
		}
		if _, err := k.HandleQueuedWithdrawInsuranceRequests(ctx); err != nil {
			panic(err)
		}
		newlyRankedInInsurances, rankOutInsurances, err := k.RankInsurances(ctx)
		if err != nil {
			panic(err)
		}
		if err = k.RePairRankedInsurances(ctx, newlyRankedInInsurances, rankOutInsurances); err != nil {
			panic(err)
		}
		k.IncrementEpoch(ctx)
	}
}
