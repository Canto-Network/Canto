package keeper_test

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.OnboardingKeeper.GetParams(suite.ctx)
	params.EnableOnboarding = false
	suite.app.OnboardingKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.OnboardingKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
