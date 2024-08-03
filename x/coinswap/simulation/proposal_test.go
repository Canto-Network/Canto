package simulation_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v8/x/coinswap/simulation"
	"github.com/Canto-Network/Canto/v8/x/coinswap/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/require"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	ctx := sdk.NewContext(nil, cmtproto.Header{}, true, nil)
	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs()
	require.Equal(t, 1, len(weightedProposalMsgs))

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightMsgUpdateParams, w0.AppParamsKey())
	require.Equal(t, simulation.DefaultWeightMsgUpdateParams, w0.DefaultWeight())

	msg := w0.MsgSimulatorFn()(r, ctx, accounts)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	require.True(t, ok)
	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgUpdateParams.Authority)
	require.Equal(t, math.LegacyNewDecWithPrec(0, 3), msgUpdateParams.Params.Fee)
	require.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, 240456), msgUpdateParams.Params.PoolCreationFee)
	require.Equal(t, math.LegacyNewDecWithPrec(0, 3), msgUpdateParams.Params.TaxRate)
	require.Equal(t, math.NewIntWithDecimal(410694, 18), msgUpdateParams.Params.MaxStandardCoinPerPool)
	require.Equal(t, sdk.NewCoins(
		sdk.NewCoin(types.UsdcIBCDenom, math.NewIntWithDecimal(89, 6)),
		sdk.NewCoin(types.UsdtIBCDenom, math.NewIntWithDecimal(22, 6)),
		sdk.NewCoin(types.EthIBCDenom, math.NewIntWithDecimal(12, 16)),
	), msgUpdateParams.Params.MaxSwapAmount)
}
