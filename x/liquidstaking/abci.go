package liquidstaking

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if k.IsEpochReached(ctx) {
		// TODO: Paired가 아닌데 Reward가 쌓여있는 상황이 있을 수 있지 않을까?
		// Reward 외적인 Balance가 잡히진 않을지...?
		// Unknown risk를 방지하기 위해 Chunk에도 Delegation reward 정산용 Address를 하나 두는 것이 좋을 듯함
		k.DistributeReward(ctx)
		k.CoverSlashingAndHandleMatureUnbondings(ctx)
		if _, err := k.HandleQueuedLiquidUnstakes(ctx); err != nil {
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
