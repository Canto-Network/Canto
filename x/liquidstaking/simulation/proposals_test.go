package simulation_test

import (
	cantoapp "github.com/Canto-Network/Canto/v6/app"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/simulation"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"math/rand"
	"testing"
	"time"
)

func TestProposalContents(t *testing.T) {
	app, ctx := createTestApp(false)

	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 10)

	getTestingValidator0(t, app, ctx, accounts)
	getTestingValidator1(t, app, ctx, accounts)

	// begin a new block
	blockTime := time.Now().UTC()
	app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  app.LastBlockHeight() + 1,
			AppHash: app.LastCommitID().Hash,
			Time:    blockTime,
		},
	})
	app.EndBlock(abci.RequestEndBlock{Height: app.LastBlockHeight() + 1})

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(app.LiquidStakingKeeper)
	require.Len(t, weightedProposalContent, 2)

	w0 := weightedProposalContent[0]
	w1 := weightedProposalContent[1]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightSimulateUpdateDynamicFeeRateProposal, w0.AppParamsKey())
	require.Equal(t, cantoapp.DefaultWeightUpdateDynamicFeeRateProposal, w0.DefaultWeight())

	// tests w1 interface:
	require.Equal(t, simulation.OpWeightSimulateUpdateMaximumDiscountRate, w1.AppParamsKey())
	require.Equal(t, cantoapp.DefaultWeightUpdateMaximumDiscountRate, w1.DefaultWeight())

	content0 := w0.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content0)

	content1 := w1.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content1)
}
