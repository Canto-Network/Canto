package keeper_test

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"time"
)

func (suite *KeeperTestSuite) TestWithdrawingInsuranceSetGet() {
	numberWithdrawingInsurances := 10
	withdrawingInsurances := GenerateWithdrawingInsurances(numberWithdrawingInsurances)
	for _, withdrawingInsurance := range withdrawingInsurances {
		suite.app.LiquidStakingKeeper.SetWithdrawingInsurance(suite.ctx, withdrawingInsurance)
	}

	for _, withdrawingInsurance := range withdrawingInsurances {
		insuranceId := withdrawingInsurance.InsuranceId
		chunkId := withdrawingInsurance.ChunkId
		completionTime := withdrawingInsurance.CompletionTime

		// Get withdrawingInsurance from the store
		result, found := suite.app.LiquidStakingKeeper.GetWithdrawingInsurance(suite.ctx, insuranceId)

		// Validation
		suite.Require().True(found)
		suite.Require().Equal(result.InsuranceId, insuranceId)
		suite.Require().Equal(result.ChunkId, chunkId)
		suite.Require().Equal(result.CompletionTime, completionTime)
	}
}

func (suite *KeeperTestSuite) TestDeleteWithdrawingInsurance() {
	numberWithdrawingInsurances := 10
	withdrawingInsurances := GenerateWithdrawingInsurances(numberWithdrawingInsurances)
	for _, withdrawingInsurance := range withdrawingInsurances {
		suite.app.LiquidStakingKeeper.SetWithdrawingInsurance(suite.ctx, withdrawingInsurance)
	}

	for _, withdrawingInsurance := range withdrawingInsurances {
		insuranceId := withdrawingInsurance.InsuranceId
		chunkId := withdrawingInsurance.ChunkId
		completionTime := withdrawingInsurance.CompletionTime

		// Get withdrawingInsurance from the store
		result, found := suite.app.LiquidStakingKeeper.GetWithdrawingInsurance(suite.ctx, insuranceId)

		// Validation
		suite.Require().True(found)
		suite.Require().Equal(result.InsuranceId, insuranceId)
		suite.Require().Equal(result.ChunkId, chunkId)
		suite.Require().Equal(result.CompletionTime, completionTime)

		// Delete withdrawingInsurance from the store
		suite.app.LiquidStakingKeeper.DeleteWithdrawingInsurance(suite.ctx, insuranceId)

		// Get withdrawingInsurance from the store
		result, found = suite.app.LiquidStakingKeeper.GetWithdrawingInsurance(suite.ctx, insuranceId)

		// Validation
		suite.Require().False(found)
	}
}

func GenerateWithdrawingInsurances(number int) []types.WithdrawingInsurance {
	withdrawingInsurances := make([]types.WithdrawingInsurance, number)
	for i := 0; i < number; i++ {
		withdrawingInsurances[i] = types.NewWithdrawingInsurance(uint64(i), uint64(i), time.Now().UTC())
	}
	return withdrawingInsurances
}
