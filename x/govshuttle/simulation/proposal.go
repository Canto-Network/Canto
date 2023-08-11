package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v7/app/params"
	"github.com/Canto-Network/Canto/v7/x/govshuttle"
	"github.com/Canto-Network/Canto/v7/x/govshuttle/keeper"
	"github.com/Canto-Network/Canto/v7/x/govshuttle/types"
)

// Simulation operation weights constants.
const (
	OpWeightSimulateLendingMarketProposal = "op_weight_lending_market_proposal"
	OpWeightSimulateTreasuryProposal      = "op_weight_treasury_proposal"
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateLendingMarketProposal,
			params.DefaultWeightLendingMarketProposal,
			SimulateLendingMarketProposal(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateTreasuryProposal,
			params.DefaultWeightRegisterERC20Proposal,
			SimulateTreasuryProposal(k),
		),
	}
}

func SimulateLendingMarketProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {

		treasuryProposalMetadata := types.TreasuryProposalMetadata{
			PropID:    1,
			Recipient: accs[0].Address.String(),
			Amount:    uint64(1000000000000000000),
			Denom:     "canto",
		}

		treasuryProposal := types.TreasuryProposal{
			Title:       simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 100),
			Metadata:    &treasuryProposalMetadata,
		}

		lendingMarketProposal := treasuryProposal.FromTreasuryToLendingMarket()
		lendingMarketProposal.Metadata.Calldatas = []string{"callData1"}

		if err := govshuttle.NewgovshuttleProposalHandler(&k)(ctx, lendingMarketProposal); err != nil {
			panic(err)
		}

		return nil
	}
}

func SimulateTreasuryProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {

		treasuryProposalMetadata := types.TreasuryProposalMetadata{
			PropID:    1,
			Recipient: accs[0].Address.String(),
			Amount:    uint64(1000000000000000000),
			Denom:     "canto",
		}

		proposal := types.TreasuryProposal{
			Title:       simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 100),
			Metadata:    &treasuryProposalMetadata,
		}

		if err := govshuttle.NewgovshuttleProposalHandler(&k)(ctx, &proposal); err != nil {
			panic(err)
		}

		return nil
	}
}
