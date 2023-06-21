package keeper

import (
	"context"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Publish events using returned values
	_, _, _, err := k.DoLiquidStake(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgLiquidStakeResponse{}, nil
}
func (k Keeper) LiquidUnstake(goCtx context.Context, msg *types.MsgLiquidUnstake) (*types.MsgLiquidUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Publish events using returned values
	_, _, err := k.QueueLiquidUnstake(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgLiquidUnstakeResponse{}, nil
}

func (k Keeper) ProvideInsurance(goCtx context.Context, msg *types.MsgProvideInsurance) (*types.MsgProvideInsuranceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Publish events using returned values
	_, err := k.DoProvideInsurance(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgProvideInsuranceResponse{}, nil
}

func (k Keeper) CancelProvideInsurance(goCtx context.Context, msg *types.MsgCancelProvideInsurance) (*types.MsgCancelProvideInsuranceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Publish events using returned values
	_, err := k.DoCancelProvideInsurance(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgCancelProvideInsuranceResponse{}, nil
}

func (k Keeper) DepositInsurance(goCtx context.Context, msg *types.MsgDepositInsurance) (*types.MsgDepositInsuranceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.DoDepositInsurance(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositInsuranceResponse{}, nil
}

func (k Keeper) WithdrawInsurance(goCtx context.Context, msg *types.MsgWithdrawInsurance) (*types.MsgWithdrawInsuranceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Publish events using returned values
	_, err := k.DoWithdrawInsurance(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawInsuranceResponse{}, nil
}

func (k Keeper) WithdrawInsuranceCommission(goCtx context.Context, msg *types.MsgWithdrawInsuranceCommission) (*types.MsgWithdrawInsuranceCommissionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.DoWithdrawInsuranceCommission(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawInsuranceCommissionResponse{}, nil
}

func (k Keeper) ClaimDiscountedReward(goCtx context.Context, msg *types.MsgClaimDiscountedReward) (*types.MsgClaimDiscountedRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.DoClaimDiscountedReward(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgClaimDiscountedRewardResponse{}, nil
}
