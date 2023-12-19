package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/Canto-Network/Canto/v7/app/params"
	"github.com/Canto-Network/Canto/v7/x/erc20/simulation"
	"github.com/stretchr/testify/require"
)

func TestProposalContents(t *testing.T) {
	app, ctx := createTestApp(t, false)

	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 10)

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(app.Erc20Keeper, app.AccountKeeper, app.BankKeeper, app.EvmKeeper, app.FeeMarketKeeper)
	require.Len(t, weightedProposalContent, 3)

	w0 := weightedProposalContent[0]
	w1 := weightedProposalContent[1]
	w2 := weightedProposalContent[2]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightSimulateRegisterCoinProposal, w0.AppParamsKey())
	require.Equal(t, params.DefaultWeightRegisterCoinProposal, w0.DefaultWeight())

	// tests w1 interface:
	require.Equal(t, simulation.OpWeightSimulateRegisterERC20Proposal, w1.AppParamsKey())
	require.Equal(t, params.DefaultWeightRegisterERC20Proposal, w1.DefaultWeight())

	// tests w2 interface:
	require.Equal(t, simulation.OpWeightSimulateToggleTokenConversionProposal, w2.AppParamsKey())
	require.Equal(t, params.DefaultWeightToggleTokenConversionProposal, w2.DefaultWeight())

	content0 := w0.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content0)

	content1 := w1.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content1)

	content2 := w1.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content2)
}
