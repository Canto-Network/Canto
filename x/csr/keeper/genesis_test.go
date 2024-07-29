package keeper_test

import (
	"time"

	"github.com/Canto-Network/Canto/v8/x/csr"
	"github.com/Canto-Network/Canto/v8/x/csr/types"
	"github.com/evmos/ethermint/tests"
)

func (suite *KeeperTestSuite) TestDefaultGenesis() {
	genState := types.DefaultGenesis()

	csr.InitGenesis(suite.ctx, suite.app.CSRKeeper, suite.app.AccountKeeper, *genState)
	exported := csr.ExportGenesis(suite.ctx, suite.app.CSRKeeper)
	suite.Require().Equal(genState, exported)
}

func (suite *KeeperTestSuite) TestImportExportGenesisEmpty() {
	_, found := suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.Require().False(found)
	csrs := suite.app.CSRKeeper.GetAllCSRs(suite.ctx)
	suite.Require().Empty(csrs)

	genState := csr.ExportGenesis(suite.ctx, suite.app.CSRKeeper)
	suite.Require().Equal("", genState.TurnstileAddress)
	suite.Require().Empty(genState.Csrs)

	// Copy genState to genState2 and init with it
	var genState2 types.GenesisState
	bz := suite.app.AppCodec().MustMarshalJSON(genState)
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	csr.InitGenesis(suite.ctx, suite.app.CSRKeeper, suite.app.AccountKeeper, genState2)

	_, found = suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.Require().False(found)
	csrs = suite.app.CSRKeeper.GetAllCSRs(suite.ctx)
	suite.Require().Empty(csrs)
	genState3 := csr.ExportGenesis(suite.ctx, suite.app.CSRKeeper)
	suite.Equal(*genState, genState2)
	suite.Equal(genState2, *genState3)
	suite.Require().Equal("", genState3.TurnstileAddress)
	suite.Require().Empty(genState3.Csrs)
}

func (suite *KeeperTestSuite) TestInitExportGenesis() {
	expGenesis := types.DefaultGenesis()

	csr.InitGenesis(suite.ctx, suite.app.CSRKeeper, suite.app.AccountKeeper, *expGenesis)
	genState := csr.ExportGenesis(suite.ctx, suite.app.CSRKeeper)
	suite.Require().Equal(expGenesis, genState)

	bz := suite.app.AppCodec().MustMarshalJSON(genState)

	var genState2 types.GenesisState
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	csr.InitGenesis(suite.ctx, suite.app.CSRKeeper, suite.app.AccountKeeper, genState2)
	genState3 := csr.ExportGenesis(suite.ctx, suite.app.CSRKeeper)

	suite.Require().Equal(*genState, genState2)
	suite.Require().Equal(genState2, *genState3)
}

func (suite *KeeperTestSuite) TestImportExportGenesis() {
	t, _ := time.Parse(time.RFC3339, "2022-01-01T00:00:00Z")
	suite.ctx = suite.ctx.WithBlockHeight(1).WithBlockTime(t)

	numberCSRs := 10
	csrs := GenerateCSRs(numberCSRs)
	for _, csr := range csrs {
		suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
	}

	turnstileAddress := tests.GenerateAddress()
	suite.app.CSRKeeper.SetTurnstile(suite.ctx, turnstileAddress)

	genState := csr.ExportGenesis(suite.ctx, suite.app.CSRKeeper)
	bz := suite.app.AppCodec().MustMarshalJSON(genState)

	// Copy genState to genState2 and init with it
	var genState2 types.GenesisState
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	csr.InitGenesis(suite.ctx, suite.app.CSRKeeper, suite.app.AccountKeeper, genState2)
	exported := csr.ExportGenesis(suite.ctx, suite.app.CSRKeeper)
	suite.Equal(*genState, *exported)

	suite.ctx = suite.ctx.WithBlockHeight(1).WithBlockTime(t)

	c := suite.app.CSRKeeper.GetAllCSRs(suite.ctx)
	suite.Equal(csrs, c)

	ta, found := suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.True(found)
	suite.Equal(turnstileAddress, ta)
}
