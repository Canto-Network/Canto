package keeper_test

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"time"
)

func (suite *KeeperTestSuite) TestRedelegationInfoSetGet() {
	t, err := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	suite.NoError(err)
	info := types.NewRedelegationInfo(1, t)
	suite.app.LiquidStakingKeeper.SetRedelegationInfo(suite.ctx, info)

	result, found := suite.app.LiquidStakingKeeper.GetRedelegationInfo(suite.ctx, 1)
	suite.True(found)
	suite.Equal(info, result)
}

func (suite *KeeperTestSuite) TestDeleteRedelegationInfo() {
	t, err := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	suite.NoError(err)
	info := types.NewRedelegationInfo(1, t)
	suite.app.LiquidStakingKeeper.SetRedelegationInfo(suite.ctx, info)

	result, found := suite.app.LiquidStakingKeeper.GetRedelegationInfo(suite.ctx, 1)
	suite.True(found)
	suite.Equal(info, result)

	suite.app.LiquidStakingKeeper.DeleteRedelegationInfo(suite.ctx, 1)

	result, found = suite.app.LiquidStakingKeeper.GetRedelegationInfo(suite.ctx, 1)
	suite.False(found)
	suite.Equal(types.RedelegationInfo{}, result)
}

func (suite *KeeperTestSuite) TestGetAllRedelegationInfos() {
	t, err := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	suite.NoError(err)
	info1 := types.NewRedelegationInfo(1, t)
	info2 := types.NewRedelegationInfo(2, t)
	suite.app.LiquidStakingKeeper.SetRedelegationInfo(suite.ctx, info1)
	suite.app.LiquidStakingKeeper.SetRedelegationInfo(suite.ctx, info2)

	result := suite.app.LiquidStakingKeeper.GetAllRedelegationInfos(suite.ctx)
	suite.Equal([]types.RedelegationInfo{info1, info2}, result)
}

func (suite *KeeperTestSuite) TestIterteAllRedelegationInfos() {
	t, err := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	suite.NoError(err)
	info1 := types.NewRedelegationInfo(1, t)
	info2 := types.NewRedelegationInfo(2, t)
	suite.app.LiquidStakingKeeper.SetRedelegationInfo(suite.ctx, info1)
	suite.app.LiquidStakingKeeper.SetRedelegationInfo(suite.ctx, info2)

	var result []types.RedelegationInfo
	suite.app.LiquidStakingKeeper.IterateAllRedelegationInfos(suite.ctx, func(info types.RedelegationInfo) bool {
		result = append(result, info)
		return false
	})
	suite.Equal([]types.RedelegationInfo{info1, info2}, result)
}

func (suite *KeeperTestSuite) TestDeleteNonExistingRedelegationInfo() {
	suite.app.LiquidStakingKeeper.DeleteRedelegationInfo(suite.ctx, 1000)
}
