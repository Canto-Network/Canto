package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/Canto-Network/Canto/v8/app/params"
	"github.com/Canto-Network/Canto/v8/x/erc20/simulation"
	"github.com/Canto-Network/Canto/v8/x/erc20/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/require"
)

func TestProposalMsgs(t *testing.T) {
	app, ctx := createTestApp(t, false)

	// initialize parameters
	s := rand.NewSource(2)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 10)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs(app.Erc20Keeper, app.AccountKeeper, app.BankKeeper, app.EvmKeeper, app.FeeMarketKeeper)
	require.Equal(t, 4, len(weightedProposalMsgs))

	w0 := weightedProposalMsgs[0]
	w1 := weightedProposalMsgs[1]
	w2 := weightedProposalMsgs[2]
	w3 := weightedProposalMsgs[3]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightMsgUpdateParams, w0.AppParamsKey())
	require.Equal(t, simulation.DefaultWeightMsgUpdateParams, w0.DefaultWeight())

	// tests w1 interface:
	require.Equal(t, simulation.OpWeightSimulateRegisterCoinProposal, w1.AppParamsKey())
	require.Equal(t, params.DefaultWeightRegisterCoinProposal, w1.DefaultWeight())

	// tests w2 interface:
	require.Equal(t, simulation.OpWeightSimulateRegisterERC20Proposal, w2.AppParamsKey())
	require.Equal(t, params.DefaultWeightRegisterERC20Proposal, w2.DefaultWeight())

	// tests w3 interface:
	require.Equal(t, simulation.OpWeightSimulateToggleTokenConversionProposal, w3.AppParamsKey())
	require.Equal(t, params.DefaultWeightToggleTokenConversionProposal, w3.DefaultWeight())

	msg := w0.MsgSimulatorFn()(r, ctx, accounts)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	require.True(t, ok)

	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgUpdateParams.Authority)
	require.Equal(t, true, msgUpdateParams.Params.EnableErc20)
	require.Equal(t, true, msgUpdateParams.Params.EnableEVMHook)

	msg = w1.MsgSimulatorFn()(r, ctx, accounts)
	msgRegisterCoin, ok := msg.(*types.MsgRegisterCoin)
	require.True(t, ok)

	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgRegisterCoin.Authority)
	require.NotNil(t, msgRegisterCoin.Title)
	require.NotNil(t, msgRegisterCoin.Description)
	require.NotNil(t, msgRegisterCoin.Metadata)

	msg = w2.MsgSimulatorFn()(r, ctx, accounts)
	msgRegisterERC20, ok := msg.(*types.MsgRegisterERC20)
	require.True(t, ok)

	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgRegisterERC20.Authority)
	require.NotNil(t, msgRegisterERC20.Title)
	require.NotNil(t, msgRegisterERC20.Description)
	require.NotNil(t, msgRegisterERC20.Erc20Address)

	msg = w3.MsgSimulatorFn()(r, ctx, accounts)
	msgToggleTokenConversion, ok := msg.(*types.MsgToggleTokenConversion)
	require.True(t, ok)

	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgToggleTokenConversion.Authority)
	require.NotNil(t, msgToggleTokenConversion.Title)
	require.NotNil(t, msgToggleTokenConversion.Description)
	require.NotNil(t, msgToggleTokenConversion.Token)
}
