package keeper_test

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestUnpairingForUnstakingChunkInfoSetGet() {
	delegator := sdk.AccAddress("delegator")
	escrowedCoin := sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100))
	info := types.NewUnpairingForUnstakingChunkInfo(1, delegator.String(), escrowedCoin)
	suite.app.LiquidStakingKeeper.SetUnpairingForUnstakingChunkInfo(suite.ctx, info)

	result, found := suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, 1)
	suite.True(found)
	suite.Equal(info, result)
}

func (suite *KeeperTestSuite) TestDeleteUnpairingForUnstakingChunkInfo() {
	delegator := sdk.AccAddress("delegator")
	escrowedCoin := sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100))
	info := types.NewUnpairingForUnstakingChunkInfo(1, delegator.String(), escrowedCoin)
	suite.app.LiquidStakingKeeper.SetUnpairingForUnstakingChunkInfo(suite.ctx, info)

	result, found := suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, 1)
	suite.True(found)
	suite.Equal(info, result)

	suite.app.LiquidStakingKeeper.DeleteUnpairingForUnstakingChunkInfo(suite.ctx, 1)

	result, found = suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, 1)
	suite.False(found)
	suite.Equal(types.UnpairingForUnstakingChunkInfo{}, result)
}

func (suite *KeeperTestSuite) TestGetAllUnpairingForUnstakingChunkInfos() {
	delegator := sdk.AccAddress("delegator")
	escrowedCoin := sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100))
	info1 := types.NewUnpairingForUnstakingChunkInfo(1, delegator.String(), escrowedCoin)
	info2 := types.NewUnpairingForUnstakingChunkInfo(2, delegator.String(), escrowedCoin)
	suite.app.LiquidStakingKeeper.SetUnpairingForUnstakingChunkInfo(suite.ctx, info1)
	suite.app.LiquidStakingKeeper.SetUnpairingForUnstakingChunkInfo(suite.ctx, info2)

	result := suite.app.LiquidStakingKeeper.GetAllUnpairingForUnstakingChunkInfos(suite.ctx)
	suite.Equal([]types.UnpairingForUnstakingChunkInfo{info1, info2}, result)
}

func (suite *KeeperTestSuite) TestIterateAllUnpairingForUnstakingChunkInfos() {
	delegator := sdk.AccAddress("delegator")
	escrowedCoin := sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100))
	info1 := types.NewUnpairingForUnstakingChunkInfo(1, delegator.String(), escrowedCoin)
	info2 := types.NewUnpairingForUnstakingChunkInfo(2, delegator.String(), escrowedCoin)
	suite.app.LiquidStakingKeeper.SetUnpairingForUnstakingChunkInfo(suite.ctx, info1)
	suite.app.LiquidStakingKeeper.SetUnpairingForUnstakingChunkInfo(suite.ctx, info2)

	var result []types.UnpairingForUnstakingChunkInfo
	suite.app.LiquidStakingKeeper.IterateAllUnpairingForUnstakingChunkInfos(suite.ctx, func(info types.UnpairingForUnstakingChunkInfo) bool {
		result = append(result, info)
		return false
	})
	suite.Equal([]types.UnpairingForUnstakingChunkInfo{info1, info2}, result)
}

func (suite *KeeperTestSuite) TestDeleteNonExistingUnpairingForUnstakingChunkInfo() {
	suite.NotPanics(
		func() { suite.app.LiquidStakingKeeper.DeleteUnpairingForUnstakingChunkInfo(suite.ctx, 1000) },
		"should not panic",
	)
}
