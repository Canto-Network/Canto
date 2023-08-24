package keeper_test

import (
	"time"

	"github.com/evmos/ethermint/tests"

	"github.com/Canto-Network/Canto/v7/x/govshuttle"
	"github.com/Canto-Network/Canto/v7/x/govshuttle/types"
)

func (suite *KeeperTestSuite) TestDefaultGenesis() {
	genState := types.DefaultGenesis()

	govshuttle.InitGenesis(suite.ctx, suite.app.GovshuttleKeeper, suite.app.AccountKeeper, *genState)
	got := govshuttle.ExportGenesis(suite.ctx, suite.app.GovshuttleKeeper)
	suite.Require().Equal(genState, got)
}

func (suite *KeeperTestSuite) TestImportExportGenesisEmpty() {
	_, found := suite.app.GovshuttleKeeper.GetPort(suite.ctx)
	suite.Require().False(found)
	genState := govshuttle.ExportGenesis(suite.ctx, suite.app.GovshuttleKeeper)
	suite.Require().Nil(genState.PortAddress)

	// Copy genState to genState2 and init with it
	var genState2 types.GenesisState
	bz := suite.app.AppCodec().MustMarshalJSON(genState)
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	govshuttle.InitGenesis(suite.ctx, suite.app.GovshuttleKeeper, suite.app.AccountKeeper, genState2)

	_, found = suite.app.GovshuttleKeeper.GetPort(suite.ctx)
	suite.Require().False(found)

	genState3 := govshuttle.ExportGenesis(suite.ctx, suite.app.GovshuttleKeeper)
	suite.Equal(*genState, genState2)
	suite.Equal(genState2, *genState3)
	suite.Require().Nil(genState.PortAddress)
}

func (suite *KeeperTestSuite) TestInitExportGenesis() {
	portAddress := tests.GenerateAddress()
	expGenesis := types.NewGenesisState(types.DefaultParams(), portAddress.Bytes())

	govshuttle.InitGenesis(suite.ctx, suite.app.GovshuttleKeeper, suite.app.AccountKeeper, *expGenesis)
	genState := govshuttle.ExportGenesis(suite.ctx, suite.app.GovshuttleKeeper)
	suite.Require().Equal(expGenesis, genState)

	bz := suite.app.AppCodec().MustMarshalJSON(genState)

	var genState2 types.GenesisState
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	govshuttle.InitGenesis(suite.ctx, suite.app.GovshuttleKeeper, suite.app.AccountKeeper, genState2)
	genState3 := govshuttle.ExportGenesis(suite.ctx, suite.app.GovshuttleKeeper)

	suite.Require().Equal(*genState, genState2)
	suite.Require().Equal(genState2, *genState3)
}

func (suite *KeeperTestSuite) TestImportExportGenesis() {
	t, _ := time.Parse(time.RFC3339, "2022-01-01T00:00:00Z")
	suite.ctx = suite.ctx.WithBlockHeight(1).WithBlockTime(t)

	portAddress := tests.GenerateAddress()
	suite.app.GovshuttleKeeper.SetPort(suite.ctx, portAddress)

	genState := govshuttle.ExportGenesis(suite.ctx, suite.app.GovshuttleKeeper)
	bz := suite.app.AppCodec().MustMarshalJSON(genState)

	// Copy genState to genState2 and init with it
	var genState2 types.GenesisState
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	govshuttle.InitGenesis(suite.ctx, suite.app.GovshuttleKeeper, suite.app.AccountKeeper, genState2)
	exported := govshuttle.ExportGenesis(suite.ctx, suite.app.GovshuttleKeeper)
	suite.Equal(*genState, *exported)

	suite.ctx = suite.ctx.WithBlockHeight(1).WithBlockTime(t)

	p, found := suite.app.GovshuttleKeeper.GetPort(suite.ctx)
	suite.True(found)
	suite.Equal(portAddress, p)
}
