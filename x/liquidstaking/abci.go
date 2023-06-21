package liquidstaking

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if k.IsEpochReached(ctx) {
		// Reward 외적인 Balance가 잡히진 않을지...?
		// Unknown risk를 방지하기 위해 Chunk에도 Delegation reward 정산용 Address를 하나 두는 것이 좋을 듯함
		// TODO: 정책적으로, 스펙레벨로 결정이 필요할 듯함
		// TODO: 지금은 이전 Epoch 스테이트를 기준으로 책정하는데 스펙 레벨에서 보면 이걸 인지하기 어려움
		k.DistributeReward(ctx)
		k.CoverSlashingAndHandleMatureUnbondings(ctx)
		if _, err := k.HandleQueuedLiquidUnstakes(ctx); err != nil {
			panic(err)
		}
		if err := k.HandleUnprocessedQueuedLiquidUnstakes(ctx); err != nil {
			panic(err)
		}
		// TODO: remaining queued liquid unstake infos should be deleted
		// - get escrowed ls tokens to liquid unstaker if info is still remained and unbonding is not started for some reason
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
