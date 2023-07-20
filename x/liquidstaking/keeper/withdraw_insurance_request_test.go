package keeper_test

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
)

func (suite *KeeperTestSuite) TestWithdrawInsuranceRequestSetGet() {
	info := types.NewWithdrawInsuranceRequest(1)
	suite.app.LiquidStakingKeeper.SetWithdrawInsuranceRequest(suite.ctx, info)

	result, found := suite.app.LiquidStakingKeeper.GetWithdrawInsuranceRequest(suite.ctx, 1)
	suite.True(found)
	suite.Equal(info, result)
}

func (suite *KeeperTestSuite) TestDeleteWithdrawInsuranceRequest() {
	info := types.NewWithdrawInsuranceRequest(1)
	suite.app.LiquidStakingKeeper.SetWithdrawInsuranceRequest(suite.ctx, info)

	result, found := suite.app.LiquidStakingKeeper.GetWithdrawInsuranceRequest(suite.ctx, 1)
	suite.True(found)
	suite.Equal(info, result)

	suite.app.LiquidStakingKeeper.DeleteWithdrawInsuranceRequest(suite.ctx, 1)

	result, found = suite.app.LiquidStakingKeeper.GetWithdrawInsuranceRequest(suite.ctx, 1)
	suite.False(found)
	suite.Equal(types.WithdrawInsuranceRequest{}, result)
}

func (suite *KeeperTestSuite) TestGetAllWithdrawInsuranceRequests() {
	info1 := types.NewWithdrawInsuranceRequest(1)
	info2 := types.NewWithdrawInsuranceRequest(2)
	suite.app.LiquidStakingKeeper.SetWithdrawInsuranceRequest(suite.ctx, info1)
	suite.app.LiquidStakingKeeper.SetWithdrawInsuranceRequest(suite.ctx, info2)

	result := suite.app.LiquidStakingKeeper.GetAllWithdrawInsuranceRequests(suite.ctx)
	suite.Equal([]types.WithdrawInsuranceRequest{info1, info2}, result)
}

func (suite *KeeperTestSuite) TestIterateAllWithdrawInsuranceRequests() {
	info1 := types.NewWithdrawInsuranceRequest(1)
	info2 := types.NewWithdrawInsuranceRequest(2)
	suite.app.LiquidStakingKeeper.SetWithdrawInsuranceRequest(suite.ctx, info1)
	suite.app.LiquidStakingKeeper.SetWithdrawInsuranceRequest(suite.ctx, info2)

	var result []types.WithdrawInsuranceRequest
	suite.app.LiquidStakingKeeper.IterateAllWithdrawInsuranceRequests(suite.ctx, func(info types.WithdrawInsuranceRequest) bool {
		result = append(result, info)
		return false
	})
	suite.Equal([]types.WithdrawInsuranceRequest{info1, info2}, result)
}

func (suite *KeeperTestSuite) TestDeleteNonExistingWithdrawInsuranceRequest() {
	suite.app.LiquidStakingKeeper.DeleteWithdrawInsuranceRequest(suite.ctx, 1000)
}
