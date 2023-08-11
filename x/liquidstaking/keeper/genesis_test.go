package keeper_test

import (
	"time"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/evmos/ethermint/types"
)

func (suite *KeeperTestSuite) TestDefaultGenesis() {
	genState := types.DefaultGenesisState()

	liquidstaking.InitGenesis(suite.ctx, suite.app.LiquidStakingKeeper, *genState)
	got := liquidstaking.ExportGenesis(suite.ctx, suite.app.LiquidStakingKeeper)
	suite.Require().Equal(genState, got)
}

func (suite *KeeperTestSuite) TestImportExportGenesisEmpty() {
	genState := liquidstaking.ExportGenesis(suite.ctx, suite.app.LiquidStakingKeeper)

	// Copy genState to genState2 and init with it
	var genState2 types.GenesisState
	bz := suite.app.AppCodec().MustMarshalJSON(genState)
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	liquidstaking.InitGenesis(suite.ctx, suite.app.LiquidStakingKeeper, genState2)

	genState3 := liquidstaking.ExportGenesis(suite.ctx, suite.app.LiquidStakingKeeper)
	suite.Equal(*genState, genState2)
	suite.Equal(genState2, *genState3)
}

// TestImportExportGenesis set some data in the genesis and check if it is exported correctly.
func (suite *KeeperTestSuite) TestImportExportGenesis() {
	t, _ := time.Parse(time.RFC3339, "2022-01-01T00:00:00Z")
	suite.ctx = suite.ctx.WithBlockHeight(1).WithBlockTime(t)

	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	// fundingAccount have enough balance to fund the account
	suite.fundAccount(suite.ctx, fundingAccount, oneChunk.Amount.Mul(sdk.NewInt(1000)).Mul(ethermint.PowerReduction))

	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		TenPercentFeeRate,
		nil,
	)

	// create providers and delegators
	accNum := 2
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, accNum, oneInsurance.Amount)
	delegators, delegatorBalances := suite.AddTestAddrsWithFunding(fundingAccount, accNum, oneChunk.Amount)

	var insurances []types.Insurance
	var chunks []types.Chunk
	for i := 0; i < accNum; i++ {
		// provide insurance
		insurance, err := suite.app.LiquidStakingKeeper.DoProvideInsurance(
			suite.ctx,
			types.NewMsgProvideInsurance(
				providers[i].String(),
				valAddrs[i].String(),
				providerBalances[i],
				TenPercentFeeRate,
			),
		)
		suite.NoError(err)
		// liquid stake
		ret, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(
			suite.ctx,
			types.NewMsgLiquidStake(
				delegators[i].String(),
				delegatorBalances[i],
			),
		)
		suite.NoError(err)
		chunks = append(chunks, ret[0])
		// Paired when liquid staking above
		insurance, found := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, insurance.Id)
		suite.True(found)
		insurances = append(insurances, insurance)
	}
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	// request withdraw insurance
	_, withdrawRequest, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			providers[1].String(),
			insurances[1].Id,
		),
	)
	suite.NoError(err)
	_, unstakingInfos, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
		suite.ctx,
		types.NewMsgLiquidUnstake(
			delegators[0].String(),
			oneChunk,
		),
	)
	suite.NoError(err)

	genState := liquidstaking.ExportGenesis(suite.ctx, suite.app.LiquidStakingKeeper)
	bz := suite.app.AppCodec().MustMarshalJSON(genState)

	// Copy genState to genState2 and init with it
	var genState2 types.GenesisState
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	liquidstaking.InitGenesis(suite.ctx, suite.app.LiquidStakingKeeper, genState2)
	exported := liquidstaking.ExportGenesis(suite.ctx, suite.app.LiquidStakingKeeper)
	suite.Equal(*genState, *exported)

	suite.ctx = suite.ctx.WithBlockHeight(1).WithBlockTime(t)
	// check chunks and insurances are exist
	for i := 0; i < accNum; i++ {
		c, found := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, chunks[i].Id)
		suite.True(found)
		suite.True(chunks[i].Equal(c))

		ins, found := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, insurances[i].Id)
		suite.True(found)
		suite.True(insurances[i].Equal(ins))
	}
	// check unstakingInfo and withdrawRequest are exist
	info, found := suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, chunks[1].Id)
	suite.True(found)
	suite.True(unstakingInfos[0].Equal(info))
	req, found := suite.app.LiquidStakingKeeper.GetWithdrawInsuranceRequest(suite.ctx, insurances[1].Id)
	suite.True(found)
	suite.True(withdrawRequest.Equal(req))
}
