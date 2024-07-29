package simulation

import (
	"math/rand"

	"github.com/Canto-Network/Canto/v8/app/params"
	"github.com/Canto-Network/Canto/v8/x/govshuttle/keeper"
	"github.com/Canto-Network/Canto/v8/x/govshuttle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightSimulateLendingMarketProposal = "op_weight_lending_market_proposal"
	OpWeightSimulateTreasuryProposal      = "op_weight_treasury_proposal"
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs(k keeper.Keeper) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightSimulateLendingMarketProposal,
			params.DefaultWeightLendingMarketProposal,
			SimulateMsgLendingMarket(k),
		),
		simulation.NewWeightedProposalMsg(
			OpWeightSimulateTreasuryProposal,
			params.DefaultWeightRegisterERC20Proposal,
			SimulateMsgTreasury(k),
		),
	}
}

func SimulateMsgLendingMarket(k keeper.Keeper) simtypes.MsgSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
		treasuryProposalMetadata := types.TreasuryProposalMetadata{
			PropID:    1,
			Recipient: accs[r.Intn(len(accs))].Address.String(),
			Amount:    uint64(simtypes.RandIntBetween(r, 0, 10000)),
			Denom:     "canto",
		}

		treasuryProposal := types.TreasuryProposal{
			Title:       simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 100),
			Metadata:    &treasuryProposalMetadata,
		}

		lendingMarketProposal := treasuryProposal.FromTreasuryToLendingMarket()
		lendingMarketProposal.Metadata.Calldatas = []string{"callData1"}

		var authority sdk.AccAddress = address.Module("gov")

		msg := &types.MsgLendingMarketProposal{
			Authority:   authority.String(),
			Title:       lendingMarketProposal.Title,
			Description: lendingMarketProposal.Description,
			Metadata:    lendingMarketProposal.Metadata,
		}

		if _, err := k.LendingMarketProposal(ctx, msg); err != nil {
			panic(err)
		}

		return msg
	}
}

func SimulateMsgTreasury(k keeper.Keeper) simtypes.MsgSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {

		treasuryProposalMetadata := types.TreasuryProposalMetadata{
			PropID:    1,
			Recipient: accs[r.Intn(len(accs))].Address.String(),
			Amount:    uint64(simtypes.RandIntBetween(r, 0, 10000)),
			Denom:     "canto",
		}

		var authority sdk.AccAddress = address.Module("gov")

		msg := &types.MsgTreasuryProposal{
			Authority:   authority.String(),
			Title:       simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 100),
			Metadata:    &treasuryProposalMetadata,
		}

		if _, err := k.TreasuryProposal(ctx, msg); err != nil {
			panic(err)
		}

		return msg
	}
}
