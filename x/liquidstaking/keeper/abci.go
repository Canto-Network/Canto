package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k Keeper) {
	// TODO: Check epoch
	k.DistributeReward(ctx)
	k.CoverSlashingAndHandleMatureUnbondings(ctx)
	if _, err := k.HandleQueuedLiquidUnstakes(ctx); err != nil {
		panic(err)
	}
	if _, err := k.HandleQueuedWithdrawInsuranceRequests(ctx); err != nil {
		panic(err)
	}
}
