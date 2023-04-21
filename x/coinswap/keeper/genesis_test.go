package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

func TestGenesisSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestInitGenesisAndExportGenesis() {
	k, ctx := suite.app.CoinswapKeeper, suite.ctx
	expGenesis := types.GenesisState{
		Params:        types.DefaultParams(),
		StandardDenom: denomStandard,
		Pool: []types.Pool{{
			Id:                types.GetPoolId(denomETH),
			StandardDenom:     denomStandard,
			CounterpartyDenom: denomETH,
			EscrowAddress:     types.GetReservePoolAddr("lpt-1").String(),
			LptDenom:          "lpt-1",
		}},
		Sequence: 2,
	}
	k.InitGenesis(suite.ctx, expGenesis)
	genState := k.ExportGenesis(ctx)
	suite.Require().Equal(expGenesis, genState)

	bz := suite.app.AppCodec().MustMarshalJSON(&genState)

	var genState2 types.GenesisState
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	k.InitGenesis(ctx, genState2)
	genState3 := k.ExportGenesis(ctx)

	suite.Require().Equal(genState, genState2)
	suite.Require().Equal(genState2, genState3)
}
