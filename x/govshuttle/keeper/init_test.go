package keeper_test

/*
func (suite *KeeperTestSuite) TestGeneral() {
	var caller, callee common.Address
	testCases := []struct {
		name     string
		malleate func()
		res      bool
	}{
		{
			"contract is not deployed",
			func() { caller, callee = common.Address{}, common.Address{} },
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest() //reset

		tc.malleate()

		if tc.res {
			mapContract := *suite.app.govshuttleKeeper.mapContractAddr // retrieve map contract

			suite.Require().Equal(common.Address{}, mapContract) //should not be deployed yet
			suite.Require().Equal(caller, callee)
		} else {
			suite.Require().Error(err)
		}

	}
}
*/
