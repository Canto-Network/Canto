package keeper_test

import (
	"github.com/Canto-Network/Canto/v7/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Test Genesis tests that the Turnstile has been deployed on genesis, and
// that the module account exists
func (suite *KeeperTestSuite) TestGenesis() {
	// turnstile should not exist before begin block
	_, found := suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.Require().False(found)

	// begin new block
	suite.Commit()

	// Get the Turnstile address, and check that there is indeed code at the address
	turnstile, found := suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.Require().True(found)

	// check that there is indeed code at this address
	acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, turnstile)
	suite.Require().True(acc.IsContract())

	// now check that the module address is correct
	csrAddr := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Require().Equal(sdk.AccAddress(types.ModuleAddress.Bytes()), csrAddr.GetAddress())
}
