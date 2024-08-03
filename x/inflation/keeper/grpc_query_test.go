package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v8/x/inflation/types"
	ethermint "github.com/evmos/ethermint/types"
)

func (suite *KeeperTestSuite) TestPeriod() {
	var (
		req    *types.QueryPeriodRequest
		expRes *types.QueryPeriodResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default period",
			func() {
				req = &types.QueryPeriodRequest{}
				expRes = &types.QueryPeriodResponse{}
			},
			true,
		},
		{
			"set period",
			func() {
				period := uint64(9)
				suite.app.InflationKeeper.SetPeriod(suite.ctx, period)
				suite.Commit()

				req = &types.QueryPeriodRequest{}
				expRes = &types.QueryPeriodResponse{Period: period}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			res, err := suite.queryClient.Period(suite.ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEpochMintProvision() {
	var (
		req    *types.QueryEpochMintProvisionRequest
		expRes *types.QueryEpochMintProvisionResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default epochMintProvision",
			func() {
				params := types.DefaultParams()
				defaultEpochMintProvision := types.CalculateEpochMintProvision(
					params,
					uint64(0),
					30,
					sdkmath.LegacyOneDec(),
				)
				req = &types.QueryEpochMintProvisionRequest{}
				expRes = &types.QueryEpochMintProvisionResponse{
					EpochMintProvision: sdk.NewDecCoinFromDec(types.DefaultInflationDenom, defaultEpochMintProvision),
				}
			},
			true,
		},
		{
			"set epochMintProvision",
			func() {
				epochMintProvision := sdkmath.LegacyNewDec(1_000_000)
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, epochMintProvision)
				suite.Commit()

				req = &types.QueryEpochMintProvisionRequest{}
				expRes = &types.QueryEpochMintProvisionResponse{EpochMintProvision: sdk.NewDecCoinFromDec(types.DefaultInflationDenom, epochMintProvision)}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			res, err := suite.queryClient.EpochMintProvision(suite.ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSkippedEpochs() {
	var (
		req    *types.QuerySkippedEpochsRequest
		expRes *types.QuerySkippedEpochsResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default skipped epochs",
			func() {
				req = &types.QuerySkippedEpochsRequest{}
				expRes = &types.QuerySkippedEpochsResponse{}
			},
			true,
		},
		{
			"set skipped epochs",
			func() {
				skippedEpochs := uint64(9)
				suite.app.InflationKeeper.SetSkippedEpochs(suite.ctx, skippedEpochs)
				suite.Commit()

				req = &types.QuerySkippedEpochsRequest{}
				expRes = &types.QuerySkippedEpochsResponse{SkippedEpochs: skippedEpochs}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			res, err := suite.queryClient.SkippedEpochs(suite.ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCirculatingSupply() {
	// Team allocation is only set on mainnet

	// Mint coins to increase supply
	mintDenom := suite.app.InflationKeeper.GetParams(suite.ctx).MintDenom
	mintCoin := sdk.NewCoin(mintDenom, sdk.TokensFromConsensusPower(int64(400_000_000), ethermint.PowerReduction))
	err := suite.app.InflationKeeper.MintCoins(suite.ctx, mintCoin)
	suite.Require().NoError(err)

	// team allocation is zero if not on mainnet
	expCirculatingSupply := sdk.NewDecCoin(mintDenom, sdk.TokensFromConsensusPower(400_000_000, ethermint.PowerReduction))

	res, err := suite.queryClient.CirculatingSupply(suite.ctx, &types.QueryCirculatingSupplyRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expCirculatingSupply, res.CirculatingSupply)
}

func (suite *KeeperTestSuite) TestQueryInflationRate() {

	// Mint coins to increase supply
	mintDenom := suite.app.InflationKeeper.GetParams(suite.ctx).MintDenom
	mintCoin := sdk.NewCoin(mintDenom, sdk.TokensFromConsensusPower(int64(400_000_000), ethermint.PowerReduction))
	err := suite.app.InflationKeeper.MintCoins(suite.ctx, mintCoin)
	suite.Require().NoError(err)

	expInflationRate := sdkmath.LegacyMustNewDecFromStr("4.076087000000000000")
	res, err := suite.queryClient.InflationRate(suite.ctx, &types.QueryInflationRateRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expInflationRate, res.InflationRate)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	expParams := types.DefaultParams()

	res, err := suite.queryClient.Params(suite.ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}
