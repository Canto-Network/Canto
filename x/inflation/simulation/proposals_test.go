package simulation_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v8/x/inflation/simulation"
	"github.com/Canto-Network/Canto/v8/x/inflation/types"
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
	require.Equal(t, sdk.DefaultBondDenom, msgUpdateParams.Params.MintDenom) //nolint:staticcheck // we're testing deprecated code here
	require.Equal(t, types.ExponentialCalculation{
		A:             sdkmath.LegacyNewDec(6122540),
		R:             sdkmath.LegacyNewDecWithPrec(56, 2),
		C:             sdkmath.LegacyZeroDec(),
		BondingTarget: sdkmath.LegacyNewDecWithPrec(7, 2),
		MaxVariance:   sdkmath.LegacyZeroDec(),
	}, msgUpdateParams.Params.ExponentialCalculation)
	require.Equal(t, types.InflationDistribution{
		StakingRewards: sdkmath.LegacyNewDecWithPrec(94, 2),
		CommunityPool:  sdkmath.LegacyNewDecWithPrec(6, 2),
	}, msgUpdateParams.Params.InflationDistribution)
	require.Equal(t, false, msgUpdateParams.Params.EnableInflation)
}
