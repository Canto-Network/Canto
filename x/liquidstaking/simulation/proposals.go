package simulation

import (
	"github.com/Canto-Network/Canto/v6/app"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"math/rand"
)

const (
	OpWeightSimulateUpdateDynamicFeeRateProposal = "op_weight_simulate_update_dynamic_fee_rate_proposal"
	OpWeightSimulateUpdateMaximumDiscountRate    = "op_weight_simulate_update_maximum_discount_rate"
)

func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateUpdateDynamicFeeRateProposal,
			app.DefaultWeightUpdateDynamicFeeRateProposal,
			SimulateUpdateDynamicFeeRateProposal(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateUpdateMaximumDiscountRate,
			app.DefaultWeightUpdateMaximumDiscountRate,
			SimulateUpdateMaximumDiscountRate(k),
		),
	}
}

// SimulateUpdateDynamicFeeRateProposal generates random update dynamic fee rate param change proposal content.
func SimulateUpdateDynamicFeeRateProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		params := k.GetParams(ctx)
		params.DynamicFeeRate = genDynamicFeeRate(r)
		k.SetParams(ctx, params)
		return nil
	}
}

// SimulateUpdateMaximumDiscountRate generates random update maximum discount rate param change proposal content.
func SimulateUpdateMaximumDiscountRate(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		params := k.GetParams(ctx)
		params.MaximumDiscountRate = genMaximumDiscountRate(r)
		k.SetParams(ctx, params)
		return nil
	}
}
