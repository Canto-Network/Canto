package simulation

import (
	"math/rand"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const (
	OpWeightSimulateUpdateDynamicFeeRateProposal = "op_weight_simulate_update_dynamic_fee_rate_proposal"
	OpWeightSimulateUpdateMaximumDiscountRate    = "op_weight_simulate_update_maximum_discount_rate"
	OpWeightSimulateAdvanceEpoch                 = "op_weight_simulate_advance_epoch"
)

func ProposalContents(
	k keeper.Keeper,
) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		//simulation.NewWeightedProposalContent(
		//	OpWeightSimulateUpdateMaximumDiscountRate,
		//	params.DefaultWeightUpdateMaximumDiscountRate,
		//	SimulateUpdateMaximumDiscountRate(k),
		//),
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
