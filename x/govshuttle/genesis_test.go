package govshuttle_test

import (
	// "encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v8/app"
	"github.com/Canto-Network/Canto/v8/x/govshuttle"
)

type GenesisTestSuite struct {
	suite.Suite //top level testing suite

	appA *app.Canto
	ctxA sdk.Context

	appB *app.Canto
	ctxB sdk.Context
}

var s *GenesisTestSuite

func TestGenesisTestSuite(t *testing.T) {
	s = new(GenesisTestSuite)
	suite.Run(t, s)
}

func (suite *GenesisTestSuite) DoSetupTest(t require.TestingT) {
	suite.appA = app.NewCanto(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		true,
		simtestutil.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
	)
	suite.ctxA = suite.appA.NewContext(true)

	suite.appB = app.NewCanto(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		true,
		simtestutil.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
	)
	suite.ctxB = suite.appB.NewContext(true)
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *GenesisTestSuite) TestGenesis() {
	testCases := []struct {
		name     string
		portAddr common.Address
		malleate func(portAddr common.Address)
	}{
		{
			"empty port contract address",
			common.Address{},
			func(_ common.Address) {},
		},
		{
			"non-empty port contract address",
			common.HexToAddress("0x648a5Aa0C4FbF2C1CF5a3B432c2766EeaF8E402d"),
			func(portAddr common.Address) {
				suite.appA.GovshuttleKeeper.SetPort(suite.ctxA, portAddr)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate(tc.portAddr)

			portAddr, _ := suite.appA.GovshuttleKeeper.GetPort(suite.ctxA)
			suite.Require().Equal(tc.portAddr, portAddr)

			genesisState := govshuttle.ExportGenesis(suite.ctxA, suite.appA.GovshuttleKeeper)

			govshuttle.InitGenesis(suite.ctxB, suite.appB.GovshuttleKeeper, suite.appB.AccountKeeper, *genesisState)
			portAddr, _ = suite.appB.GovshuttleKeeper.GetPort(suite.ctxB)
			suite.Require().Equal(tc.portAddr, portAddr)
		})
	}
}
