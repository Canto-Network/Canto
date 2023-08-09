package simulation

import (
	"math/rand"

	"github.com/Canto-Network/Canto/v7/app/params"
	"github.com/Canto-Network/Canto/v7/x/govshuttle/keeper"
	"github.com/Canto-Network/Canto/v7/x/govshuttle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants.
const (
	OpWeightSimulateLendingMarketProposal = "op_weight_lending_market_proposal"
	OpWeightSimulateTreasuryProposal      = "op_weight_treasury_proposal"
)

// ProposalContents defines the module weighted proposals' contents for mocking param changes, other actions with keeper
func ProposalContents(ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, gk types.GovKeeper, k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateLendingMarketProposal,
			params.DefaultWeightSimulateLendingMarketProposal,
			SimulateLendingMarketProposal(sk, k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateTreasuryProposal,
			params.DefaultWeightSimulateTreasuryProposal,
			SimulateTreasuryProposal(sk, k),
		),
	}
}

// SimulateAddWhitelistValidatorsProposal generates random add whitelisted validator param change proposal content.
func SimulateLendingMarketProposal(sk types.StakingKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {

		return types.NewLendingMarketProposal()
	}
}

// SimulateUpdateWhitelistValidatorsProposal generates random update whitelisted validator param change proposal content.
func SimulateTreasuryProposal(sk types.StakingKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		return types.NewTreasuryProposal()
	}
}
