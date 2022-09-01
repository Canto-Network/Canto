package keeper_test

import (
	_ "github.com/Canto-Network/Canto/v2/x/csr/keeper"
)



// Test Address derivation with CREATE / CREATE2
func (suite *KeeperTestSuite) TestAddressDerivation() {
	type testArgs struct {
		
	}
	testCases := []struct{
		name string
	}{
		{
			"len of nonces / salts incorrect - fail",
		},
		{
			""
		}
	}
}		

// Deploy Contract, check that derived address is correct
func (suite *KeeperTestSuite) DeployContract() {

}

// Get Nonce is a testing function for use to dermine the contract nonce 
func (suite *KeeperTestSuite) GetNonce() { 

}