package keeper

import (
	"context"

	"github.com/Canto-Network/Canto/v7/x/govshuttle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (k Keeper) LendingMarketProposal(goCtx context.Context, req *types.MsgLendingMarketProposal) (*types.MsgLendingMarketProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	legacyProposal := types.LendingMarketProposal{
		Title:       req.Title,
		Description: req.Description,
		Metadata:    req.Metadata,
	}

	_, err := k.AppendLendingMarketProposal(ctx, &legacyProposal)
	if err != nil {
		return nil, err
	}

	return &types.MsgLendingMarketProposalResponse{}, nil
}

func (k Keeper) TreasuryProposal(goCtx context.Context, req *types.MsgTreasuryProposal) (*types.MsgTreasuryProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	legacyTreasuryProposal := types.TreasuryProposal{
		Title:       req.Title,
		Description: req.Description,
		Metadata:    req.Metadata,
	}

	legacyProposal := legacyTreasuryProposal.FromTreasuryToLendingMarket()

	_, err := k.AppendLendingMarketProposal(ctx, legacyProposal)
	if err != nil {
		return nil, err
	}

	return &types.MsgTreasuryProposalResponse{}, nil
}
