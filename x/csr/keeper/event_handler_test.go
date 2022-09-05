package keeper_test

import (
	_ "github.com/Canto-Network/Canto/v2/x/csr/keeper" 
)

// test failure in the case that a withdrawal event has been received for a CSR that does not exist
// test failure in the case that a withdrawal event has been received with a recipient that does not exist
// test withdrawing zero return value
// test withdrawing positive rewards value
func (suite *KeeperTestSuite) TestWithdrawalEvent() {

}

// generate event creates data field for arbitrary transaction