package simulation

import (
	"math/rand"

	"github.com/Canto-Network/Canto/v7/app/params"
	"github.com/Canto-Network/Canto/v7/x/erc20/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgUpdateParams int = 100

	OpWeightMsgUpdateParams                       = "op_weight_msg_update_params"
	OpWeightSimulateRegisterCoinProposal          = "op_weight_register_coin_proposal"
	OpWeightSimulateRegisterERC20Proposal         = "op_weight_register_erc20_proposal"
	OpWeightSimulateToggleTokenConversionProposal = "op_weight_toggle_token_conversion_proposal"
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
		simulation.NewWeightedProposalMsg(
			OpWeightSimulateRegisterCoinProposal,
			params.DefaultWeightRegisterCoinProposal,
			SimulateMsgRegisterCoin,
		),
		simulation.NewWeightedProposalMsg(
			OpWeightSimulateRegisterERC20Proposal,
			params.DefaultWeightRegisterERC20Proposal,
			SimulateMsgRegisterERC20,
		),
		simulation.NewWeightedProposalMsg(
			OpWeightSimulateToggleTokenConversionProposal,
			params.DefaultWeightToggleTokenConversionProposal,
			SimulateMsgToggleTokenConversion,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()

	params.EnableErc20 = generateRandomBool(r)
	params.EnableEVMHook = generateRandomBool(r)

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}

func SimulateMsgRegisterCoin(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	return &types.MsgRegisterCoin{
		Authority:   authority.String(),
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		Metadata:    types.GenRandomCoinMetadata(r),
	}
}

func SimulateMsgRegisterERC20(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	return &types.MsgRegisterERC20{
		Authority:    authority.String(),
		Title:        simtypes.RandStringOfLength(r, 10),
		Description:  simtypes.RandStringOfLength(r, 100),
		Erc20Address: "",
	}
}

func SimulateMsgToggleTokenConversion(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	return &types.MsgToggleTokenConversion{
		Authority:   authority.String(),
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		Token:       "",
	}
}
