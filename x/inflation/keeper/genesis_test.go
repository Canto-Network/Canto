package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// check calculated epochMintProvision at genesis
	epochMintProvision, _ := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	expMintProvision := sdk.ZeroDec()
	suite.Require().Equal(expMintProvision, epochMintProvision)
}
