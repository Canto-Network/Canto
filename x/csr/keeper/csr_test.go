package keeper_test

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/tests"
)

// set CSR, then retrieve getDeployer, getContracts / getCSR to verify that all fields are indeed the same
func (suite *KeeperTestSuite) TestCSRSetGet() {
	// set the CSR object in keeper's state,
	poolAddress := sdk.AccAddress(tests.GenerateAddress().Bytes())
	deployer := sdk.AccAddress(tests.GenerateAddress().Bytes())

	csrPool := types.NewCSRPool(
		1,
		poolAddress.String(),
	)

	contracts, _ := generateAddresses(5)

	csr := types.NewCSR(
		deployer.String(),
		contracts,
		&csrPool,
	)
	// set csr in state
	suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
	csrRet, found := suite.app.CSRKeeper.GetCSR(suite.ctx, poolAddress)
	suite.Require().True(found)
	// retrieve CSR object from store, and check that values retrieved are correct
	suite.Require().Equal(csrRet.Deployer, deployer.String())
}
