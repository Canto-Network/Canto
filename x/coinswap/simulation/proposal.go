package simulation

import (
	"math/rand"

	math "cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v8/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgUpdateParams int = 100

	OpWeightMsgUpdateParams = "op_weight_msg_update_params"
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()
	params.Fee = math.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 10)), 3)
	params.PoolCreationFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simtypes.RandIntBetween(r, 0, 1000000)))
	params.TaxRate = math.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 10)), 3)
	params.MaxStandardCoinPerPool = math.NewIntWithDecimal(int64(simtypes.RandIntBetween(r, 0, 1000000)), 18)
	params.MaxSwapAmount = sdk.NewCoins(
		sdk.NewCoin(types.UsdcIBCDenom, math.NewIntWithDecimal(int64(simtypes.RandIntBetween(r, 1, 100)), 6)),
		sdk.NewCoin(types.UsdtIBCDenom, math.NewIntWithDecimal(int64(simtypes.RandIntBetween(r, 1, 100)), 6)),
		sdk.NewCoin(types.EthIBCDenom, math.NewIntWithDecimal(int64(simtypes.RandIntBetween(r, 1, 100)), 16)),
	)

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}
