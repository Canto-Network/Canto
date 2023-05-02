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

func (k Keeper) InsuranceProvide(goCtx context.Context, msg *types.MsgInsuranceProvide) (*types.MsgInsuranceProvideResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Publish events using returned values
	_, err := k.DoInsuranceProvide(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgInsuranceProvideResponse{}, nil
}

func (k Keeper) CancelInsuranceProvide(goCtx context.Context, msg *types.MsgCancelInsuranceProvide) (*types.MsgCancelInsuranceProvideResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Publish events using returned values
	_, err := k.DoCancelInsuranceProvide(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgCancelInsuranceProvideResponse{}, nil
}

func (k Keeper) DepositInsurance(goCtx context.Context, msg *types.MsgDepositInsurance) (*types.MsgDepositInsuranceResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	panic("implement me")
}

func (k Keeper) WithdrawInsurance(goCtx context.Context, msg *types.MsgWithdrawInsurance) (*types.MsgWithdrawInsuranceResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	panic("implement me")
}

func (k Keeper) WithdrawInsuranceCommission(goCtx context.Context, msg *types.MsgWithdrawInsuranceCommission) (*types.MsgWithdrawInsuranceCommissionResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	panic("implement me")
}
