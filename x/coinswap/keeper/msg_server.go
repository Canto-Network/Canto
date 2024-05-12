package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the coinswap MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AddLiquidity(goCtx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	if err := types.ValidateMaxToken(msg.MaxToken); err != nil {
		return nil, err
	}

	if err := types.ValidateExactStandardAmt(msg.ExactStandardAmt); err != nil {
		return nil, err
	}

	if err := types.ValidateMinLiquidity(msg.MinLiquidity); err != nil {
		return nil, err
	}

	if err := types.ValidateDeadline(msg.Deadline); err != nil {
		return nil, err
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	// check that deadline has not passed
	if ctx.BlockHeader().Time.After(time.Unix(msg.Deadline, 0)) {
		return nil, errorsmod.Wrap(types.ErrInvalidDeadline, "deadline has passed for MsgAddLiquidity")
	}

	mintToken, err := m.Keeper.AddLiquidity(ctx, msg)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	)

	return &types.MsgAddLiquidityResponse{
		MintToken: &mintToken,
	}, nil
}

func (m msgServer) RemoveLiquidity(goCtx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	if err := types.ValidateMinToken(msg.MinToken); err != nil {
		return nil, err
	}

	if err := types.ValidateWithdrawLiquidity(msg.WithdrawLiquidity); err != nil {
		return nil, err
	}

	if err := types.ValidateMinStandardAmt(msg.MinStandardAmt); err != nil {
		return nil, err
	}

	if err := types.ValidateDeadline(msg.Deadline); err != nil {
		return nil, err
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	// check that deadline has not passed
	if ctx.BlockHeader().Time.After(time.Unix(msg.Deadline, 0)) {
		return nil, errorsmod.Wrap(types.ErrInvalidDeadline, "deadline has passed for MsgRemoveLiquidity")
	}
	withdrawCoins, err := m.Keeper.RemoveLiquidity(ctx, msg)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	)

	var coins = make([]*sdk.Coin, 0, withdrawCoins.Len())
	for _, coin := range withdrawCoins {
		coins = append(coins, &coin)
	}

	return &types.MsgRemoveLiquidityResponse{
		WithdrawCoins: coins,
	}, nil
}

func (m msgServer) SwapCoin(goCtx context.Context, msg *types.MsgSwapOrder) (*types.MsgSwapCoinResponse, error) {
	if err := types.ValidateInput(msg.Input); err != nil {
		return nil, err
	}

	if err := types.ValidateOutput(msg.Output); err != nil {
		return nil, err
	}

	if msg.Input.Coin.Denom == msg.Output.Coin.Denom {
		return nil, errorsmod.Wrap(types.ErrEqualDenom, "invalid swap")
	}

	if err := types.ValidateDeadline(msg.Deadline); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	// check that deadline has not passed
	if ctx.BlockHeader().Time.After(time.Unix(msg.Deadline, 0)) {
		return nil, errorsmod.Wrap(types.ErrInvalidDeadline, "deadline has passed for MsgSwapOrder")
	}

	if m.Keeper.blockedAddrs[msg.Output.Address] {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive external funds", msg.Output.Address)
	}

	if err := m.Keeper.Swap(ctx, msg); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Input.Address),
		),
	)
	return &types.MsgSwapCoinResponse{}, nil
}

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, req.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
