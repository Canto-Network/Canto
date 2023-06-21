package keeper_test

import (
	types "github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestCalcFormulaBetweenSoftCapAndOptimal() {
	for _, tc := range []struct {
		name       string
		setupParam func(params *types.DynamicFeeRate)
		u          sdk.Dec
		expected   string
	}{
		{
			"default params and u = 6%(=0.06)",
			func(params *types.DynamicFeeRate) {},
			sdk.MustNewDecFromStr("0.06"),
			"0.025000000000000000",
		},
	} {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.DefaultParams().DynamicFeeRate
			tc.setupParam(&params)
			suite.Equal(
				tc.expected,
				suite.app.LiquidStakingKeeper.CalcFormulaBetweenSoftCapAndOptimal(
					params.R0, params.USoftCap, params.UOptimal, params.Slope1, tc.u,
				).String(),
			)
		})
	}

}

func (suite *KeeperTestSuite) TestCalcFormulaUpperOptimal() {
	for _, tc := range []struct {
		name       string
		setupParam func(params *types.DynamicFeeRate)
		u          sdk.Dec
		expected   string
	}{
		{
			"default params and u = 10%(=0.1)",
			func(params *types.DynamicFeeRate) {},
			sdk.MustNewDecFromStr("0.1"),
			"0.500000000000000000",
		},
	} {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.DefaultParams().DynamicFeeRate
			tc.setupParam(&params)
			suite.Equal(
				tc.expected,
				suite.app.LiquidStakingKeeper.CalcFormulaUpperOptimal(
					params.R0, params.UOptimal, params.UHardCap, params.Slope1, params.Slope2, tc.u,
				).String(),
			)
		})
	}
}
