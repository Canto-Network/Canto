package simulation

import (
	"math/rand"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v6/app/params"
	"github.com/Canto-Network/Canto/v6/x/erc20/keeper"
)

// Simulation operation weights constants.
const (
	OpWeightSimulateRegisterCoinProposal  = "op_weight_register_coin_proposal"
	OpWeightSimulateRegisterERC20Proposal = "op_weight_register_erc20_proposal"
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateRegisterCoinProposal,
			params.DefaultWeightRegisterCoinProposal,
			SimulateRegisterCoinProposal(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateRegisterERC20Proposal,
			params.DefaultWeightRegisterERC20Proposal,
			SimulateRegisterERC20Proposal(k),
		),
	}
}

func SimulateRegisterCoinProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		coinMetadata := genRandomCoinMetadata(r)
		params := k.GetParams(ctx)
		params.EnableErc20 = true
		k.SetParams(ctx, params)
		k.RegisterCoin(ctx, coinMetadata)
		return nil
	}
}

func SimulateRegisterERC20Proposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		genRandomCoinMetadata(r)
		return nil
	}
}

func genRandomCoinMetadata(r *rand.Rand) banktypes.Metadata {
	randDescription := simtypes.RandStringOfLength(r, 10)
	randTokenBase := "a" + simtypes.RandStringOfLength(r, 4)
	randSymbol := strings.ToUpper(simtypes.RandStringOfLength(r, 4))

	validMetadata := banktypes.Metadata{
		Description: randDescription,
		Base:        randTokenBase,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    randTokenBase,
				Exponent: 0,
			},
			{
				Denom:    randTokenBase[1:],
				Exponent: uint32(18),
			},
		},
		Name:    randTokenBase,
		Symbol:  randSymbol,
		Display: randTokenBase,
	}

	return validMetadata
}
