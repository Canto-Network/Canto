package keeper

import (
	"context"
	"strconv"
	"strings"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	chunks, totalNewShares, totalLsTokenMintAmount, err := k.DoLiquidStake(ctx, msg)
	if err != nil {
		return nil, err
	}
	var chunkIds []string
	chunkIds = []string{}
	for _, chunk := range chunks {
		chunkIds = append(chunkIds, strconv.FormatUint(chunk.Id, 10))
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgLiquidStake,
			sdk.NewAttribute(types.AttributeKeyChunkIds, strings.Join(chunkIds, ", ")),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, totalNewShares.String()),
			sdk.NewAttribute(
				types.AttributeKeyLsTokenMintedAmount,
				sdk.Coin{Denom: types.DefaultLiquidBondDenom, Amount: totalLsTokenMintAmount}.String(),
			),
		),
	})
	return &types.MsgLiquidStakeResponse{}, nil
}

func (k Keeper) LiquidUnstake(goCtx context.Context, msg *types.MsgLiquidUnstake) (*types.MsgLiquidUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, infos, err := k.QueueLiquidUnstake(ctx, msg)
	if err != nil {
		return nil, err
	}
	var toBeUnstakedChunkIds []string
	escrowedLsTokens := sdk.Coins{}
	toBeUnstakedChunkIds = []string{}
	for _, info := range infos {
		toBeUnstakedChunkIds = append(toBeUnstakedChunkIds, strconv.FormatUint(info.ChunkId, 10))
		escrowedLsTokens = escrowedLsTokens.Add(info.EscrowedLstokens)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgLiquidUnstake,
			sdk.NewAttribute(types.AttributeKeyChunkIds, strings.Join(toBeUnstakedChunkIds, ", ")),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(
				types.AttributeKeyEscrowedLsTokens,
				escrowedLsTokens.String(),
			),
		),
	})

	return &types.MsgLiquidUnstakeResponse{}, nil
}

func (k Keeper) ProvideInsurance(goCtx context.Context, msg *types.MsgProvideInsurance) (*types.MsgProvideInsuranceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	insurance, err := k.DoProvideInsurance(ctx, msg)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgProvideInsurance,
			sdk.NewAttribute(types.AttributeKeyInsuranceId, strconv.FormatUint(insurance.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyInsuranceProvider, msg.ProviderAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
	})

	return &types.MsgProvideInsuranceResponse{}, nil
}

func (k Keeper) CancelProvideInsurance(goCtx context.Context, msg *types.MsgCancelProvideInsurance) (*types.MsgCancelProvideInsuranceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	canceledInsurance, err := k.DoCancelProvideInsurance(ctx, msg)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgCancelProvideInsurance,
			sdk.NewAttribute(types.AttributeKeyInsuranceId, strconv.FormatUint(canceledInsurance.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyInsuranceProvider, msg.ProviderAddress),
		),
	})

	return &types.MsgCancelProvideInsuranceResponse{}, nil
}

func (k Keeper) DepositInsurance(goCtx context.Context, msg *types.MsgDepositInsurance) (*types.MsgDepositInsuranceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.DoDepositInsurance(ctx, msg)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgDepositInsurance,
			sdk.NewAttribute(types.AttributeKeyInsuranceId, strconv.FormatUint(msg.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyInsuranceProvider, msg.ProviderAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
	})

	return &types.MsgDepositInsuranceResponse{}, nil
}

func (k Keeper) WithdrawInsurance(goCtx context.Context, msg *types.MsgWithdrawInsurance) (*types.MsgWithdrawInsuranceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	insurance, request, err := k.DoWithdrawInsurance(ctx, msg)
	if err != nil {
		return nil, err
	}
	// If the request is queued, it means withdraw is not started yet.
	// In queued, withdrawal process will be started at upcoming epoch.
	queued := request.Equal(types.WithdrawInsuranceRequest{})
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgWithdrawInsurance,
			sdk.NewAttribute(types.AttributeKeyInsuranceId, strconv.FormatUint(insurance.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyInsuranceProvider, msg.ProviderAddress),
			sdk.NewAttribute(types.AttributeKeyWithdrawInsuranceRequestQueued, strconv.FormatBool(queued)),
		),
	})
	return &types.MsgWithdrawInsuranceResponse{}, nil
}

func (k Keeper) WithdrawInsuranceCommission(goCtx context.Context, msg *types.MsgWithdrawInsuranceCommission) (*types.MsgWithdrawInsuranceCommissionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	balances, err := k.DoWithdrawInsuranceCommission(ctx, msg)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgWithdrawInsuranceCommission,
			sdk.NewAttribute(types.AttributeKeyInsuranceId, strconv.FormatUint(msg.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyInsuranceProvider, msg.ProviderAddress),
			sdk.NewAttribute(types.AttributeKeyWithdrawnInsuranceCommission, balances.String()),
		),
	})

	return &types.MsgWithdrawInsuranceCommissionResponse{}, nil
}

func (k Keeper) ClaimDiscountedReward(goCtx context.Context, msg *types.MsgClaimDiscountedReward) (*types.MsgClaimDiscountedRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	claim, discountedMintRate, err := k.DoClaimDiscountedReward(ctx, msg)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeMsgClaimDiscountedReward,
			sdk.NewAttribute(types.AttributeKeyRequester, msg.RequesterAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyClaimTokens, claim.String()),
			sdk.NewAttribute(types.AttributeKeyDiscountedMintRate, discountedMintRate.String()),
		),
	})
	return &types.MsgClaimDiscountedRewardResponse{}, nil
}
