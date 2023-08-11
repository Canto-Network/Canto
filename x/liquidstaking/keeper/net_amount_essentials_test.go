package keeper_test

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGetNetAmountState_TotalRemainingRewards() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(40),
	})

	suite.ctx = suite.advanceHeight(suite.ctx, 100, "delegation rewards are accumulated")
	nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)

	cachedCtx, _ := suite.ctx.CacheContext()
	suite.app.DistrKeeper.WithdrawDelegationRewards(cachedCtx, env.pairedChunks[0].DerivedAddress(), env.insurances[0].GetValidator())
	delReward := suite.app.BankKeeper.GetBalance(cachedCtx, env.pairedChunks[0].DerivedAddress(), suite.denom)
	totalDelReward := delReward.Amount.MulRaw(int64(len(env.pairedChunks)))
	suite.Equal("8999964000143999250000", totalDelReward.String())

	// Calc TotalRemainingRewards manually
	rest := totalDelReward.ToDec().Mul(sdk.OneDec().Sub(TenPercentFeeRate))
	remaining := rest.Mul(sdk.OneDec().Sub(nase.FeeRate))
	result := remaining.Mul(sdk.OneDec().Sub(nase.DiscountRate))
	suite.Equal("7578851328645878416158.763952739771150000", result.String())
	suite.Equal(result.String(), nase.TotalRemainingRewards.String())
	suite.True(
		totalDelReward.GT(nase.TotalRemainingRewards.TruncateInt()),
		"total del reward should be greater than total remaining rewards",
	)
}
