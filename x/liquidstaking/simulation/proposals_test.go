package simulation_test

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/simulation"
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
	require.Len(t, weightedProposalContent, 0)
}
