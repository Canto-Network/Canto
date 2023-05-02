package keeper_test

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/tests"
	"time"
)

func (suite *KeeperTestSuite) TestLiquidUnstakeUnbondingDelegationInfoSetGet() {
	numberLiquidUnstakeUnbondingDelegationInfos := 10
	liquidUnstakeUnbondingDelegationInfos := GenerateLiquidUnstakeUnbondingDelegationInfos(numberLiquidUnstakeUnbondingDelegationInfos)
	for _, liquidUnstakeUnbondingDelegationInfo := range liquidUnstakeUnbondingDelegationInfos {
		suite.app.LiquidStakingKeeper.SetLiquidUnstakeUnbondingDelegationInfo(suite.ctx, liquidUnstakeUnbondingDelegationInfo)
	}

	for _, liquidUnstakeUnbondingDelegationInfo := range liquidUnstakeUnbondingDelegationInfos {
		chunkId := liquidUnstakeUnbondingDelegationInfo.ChunkId
		delegatorAddress := liquidUnstakeUnbondingDelegationInfo.DelegatorAddress
		validatorAddress := liquidUnstakeUnbondingDelegationInfo.ValidatorAddress
		burnAmount := liquidUnstakeUnbondingDelegationInfo.BurnAmount
		completionTime := liquidUnstakeUnbondingDelegationInfo.CompletionTime

		// Get liquidUnstakeUnbondingDelegationInfo from the store
		result, found := suite.app.LiquidStakingKeeper.GetLiquidUnstakeUnbondingDelegationInfo(suite.ctx, chunkId)

		// Validation
		suite.Require().True(found)
		suite.Require().Equal(result.ChunkId, chunkId)
		suite.Require().Equal(result.DelegatorAddress, delegatorAddress)
		suite.Require().Equal(result.ValidatorAddress, validatorAddress)
		suite.Require().Equal(result.BurnAmount, burnAmount)
		suite.Require().Equal(result.CompletionTime, completionTime)
	}
}

func (suite *KeeperTestSuite) TestDeleteLiquidUnstakeUnbondingDelegationInfo() {
	numberLiquidUnstakeUnbondingDelegationInfos := 10
	liquidUnstakeUnbondingDelegationInfos := GenerateLiquidUnstakeUnbondingDelegationInfos(numberLiquidUnstakeUnbondingDelegationInfos)
	for _, liquidUnstakeUnbondingDelegationInfo := range liquidUnstakeUnbondingDelegationInfos {
		suite.app.LiquidStakingKeeper.SetLiquidUnstakeUnbondingDelegationInfo(suite.ctx, liquidUnstakeUnbondingDelegationInfo)
	}

	for _, liquidUnstakeUnbondingDelegationInfo := range liquidUnstakeUnbondingDelegationInfos {
		chunkId := liquidUnstakeUnbondingDelegationInfo.ChunkId
		// Get liquidUnstakeUnbondingDelegationInfo from the store
		result, found := suite.app.LiquidStakingKeeper.GetLiquidUnstakeUnbondingDelegationInfo(suite.ctx, chunkId)

		// Validation
		suite.Require().True(found)
		suite.Require().Equal(result.ChunkId, chunkId)

		// Delete liquidUnstakeUnbondingDelegationInfo from the store
		suite.app.LiquidStakingKeeper.DeleteLiquidUnstakeUnbondingDelegationInfo(suite.ctx, chunkId)

		// Get liquidUnstakeUnbondingDelegationInfo from the store
		result, found = suite.app.LiquidStakingKeeper.GetLiquidUnstakeUnbondingDelegationInfo(suite.ctx, chunkId)

		// Validation
		suite.Require().False(found)
		suite.Require().Equal(result.ChunkId, uint64(0))
	}
}

func GenerateLiquidUnstakeUnbondingDelegationInfos(number int) []types.LiquidUnstakeUnbondingDelegationInfo {
	liquidUnstakeUnbondingDelegationInfos := make([]types.LiquidUnstakeUnbondingDelegationInfo, number)
	for i := 0; i < number; i++ {
		liquidUnstakeUnbondingDelegationInfos[i] = types.NewLiquidUnstakeUnbondingDelegationInfo(
			uint64(i),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.NewCoin("stake", sdk.NewInt(int64(i))),
			time.Now().UTC(),
		)
	}
	return liquidUnstakeUnbondingDelegationInfos
}
