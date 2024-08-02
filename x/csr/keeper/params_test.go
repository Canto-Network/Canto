package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	_ "github.com/Canto-Network/Canto/v7/x/csr/keeper"
)

// params test suite
func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.CSRKeeper.GetDefaultParams()
	// CSR is disabled by default
	suite.Require().False(params.EnableCsr)
	// Default CSRShares are 20%
	suite.Require().Equal(params.CsrShares, sdkmath.LegacyNewDecWithPrec(20, 2))
}
