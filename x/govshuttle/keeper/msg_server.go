package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Canto-Network/Canto/v7/x/govshuttle/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k Keeper) LendingMarketProposal(ctx context.Context, req *types.MsgLendingMarketProposal) (*types.MsgLendingMarketProposalResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, err := k.AppendLendingMarketProposal(sdkCtx, req)
	if err != nil {
		return nil, err
	}

	return &types.MsgLendingMarketProposalResponse{}, nil
}

func (k Keeper) TreasuryProposal(ctx context.Context, req *types.MsgTreasuryProposal) (*types.MsgTreasuryProposalResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, err := k.AppendLendingMarketProposal(sdkCtx, req.FromTreasuryToLendingMarket())
	if err != nil {
		return nil, err
	}

	return &types.MsgTreasuryProposalResponse{}, nil
}
