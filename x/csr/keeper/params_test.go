package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"	
	_ "github.com/Canto-Network/Canto/v2/x/csr/keeper"
)

// params test suite
func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.CSRKeeper.GetParams(suite.ctx)
	// CSR is disabled by default
	suite.Require().False(params.EnableCsr)
	// Default CSRShares are 50%
	suite.Require().Equal(params.CsrShares, sdk.NewDecWithPrec(50,2))
	//Default Address Derivation Cost Create
	suite.Require().Equal(params.AddressDerivationCostCreate, uint64(50))
}
