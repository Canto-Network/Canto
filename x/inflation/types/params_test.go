package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
)

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestParamKeyTable() {
	suite.Require().IsType(paramtypes.KeyTable{}, ParamKeyTable())
}

func (suite *ParamsTestSuite) TestParamsValidate() {
	validExponentialCalculation := ExponentialCalculation{
		A:             sdkmath.LegacyNewDec(int64(16_304_348)),
		R:             sdkmath.LegacyNewDecWithPrec(35, 2),
		C:             sdkmath.LegacyZeroDec(),
		BondingTarget: sdkmath.LegacyNewDecWithPrec(66, 2),
		MaxVariance:   sdkmath.LegacyZeroDec(),
	}

	validInflationDistribution := InflationDistribution{
		StakingRewards: sdkmath.LegacyNewDecWithPrec(1000000, 6),
		CommunityPool:  sdkmath.LegacyZeroDec(),
	}

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"default",
			DefaultParams(),
			false,
		},
		{
			"valid",
			NewParams(
				"acanto",
				validExponentialCalculation,
				validInflationDistribution,
				true,
			),
			false,
		},
		{
			"valid param literal",
			Params{
				MintDenom:              "acanto",
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				EnableInflation:        true,
			},
			false,
		},
		{
			"invalid - denom",
			NewParams(
				"/acanto",
				validExponentialCalculation,
				validInflationDistribution,
				true,
			),
			true,
		},
		{
			"invalid - denom",
			Params{
				MintDenom:              "",
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				EnableInflation:        true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative A",
			Params{
				MintDenom: "acanto",
				ExponentialCalculation: ExponentialCalculation{
					A:             sdkmath.LegacyNewDec(int64(-1)),
					R:             sdkmath.LegacyNewDecWithPrec(5, 1),
					C:             sdkmath.LegacyNewDec(int64(9_375_000)),
					BondingTarget: sdkmath.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   sdkmath.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - R greater than 1",
			Params{
				MintDenom: "acanto",
				ExponentialCalculation: ExponentialCalculation{
					A:             sdkmath.LegacyNewDec(int64(300_000_000)),
					R:             sdkmath.LegacyNewDecWithPrec(5, 0),
					C:             sdkmath.LegacyNewDec(int64(9_375_000)),
					BondingTarget: sdkmath.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   sdkmath.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative R",
			Params{
				MintDenom: "acanto",
				ExponentialCalculation: ExponentialCalculation{
					A:             sdkmath.LegacyNewDec(int64(300_000_000)),
					R:             sdkmath.LegacyNewDecWithPrec(-5, 1),
					C:             sdkmath.LegacyNewDec(int64(9_375_000)),
					BondingTarget: sdkmath.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   sdkmath.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative C",
			Params{
				MintDenom: "acanto",
				ExponentialCalculation: ExponentialCalculation{
					A:             sdkmath.LegacyNewDec(int64(300_000_000)),
					R:             sdkmath.LegacyNewDecWithPrec(5, 1),
					C:             sdkmath.LegacyNewDec(int64(-9_375_000)),
					BondingTarget: sdkmath.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   sdkmath.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - BondingTarget greater than 1",
			Params{
				MintDenom: "acanto",
				ExponentialCalculation: ExponentialCalculation{
					A:             sdkmath.LegacyNewDec(int64(300_000_000)),
					R:             sdkmath.LegacyNewDecWithPrec(5, 1),
					C:             sdkmath.LegacyNewDec(int64(9_375_000)),
					BondingTarget: sdkmath.LegacyNewDecWithPrec(50, 1),
					MaxVariance:   sdkmath.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative BondingTarget",
			Params{
				MintDenom: "acanto",
				ExponentialCalculation: ExponentialCalculation{
					A:             sdkmath.LegacyNewDec(int64(300_000_000)),
					R:             sdkmath.LegacyNewDecWithPrec(5, 1),
					C:             sdkmath.LegacyNewDec(int64(9_375_000)),
					BondingTarget: sdkmath.LegacyNewDecWithPrec(50, 2).Neg(),
					MaxVariance:   sdkmath.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative max Variance",
			Params{
				MintDenom: "acanto",
				ExponentialCalculation: ExponentialCalculation{
					A:             sdkmath.LegacyNewDec(int64(300_000_000)),
					R:             sdkmath.LegacyNewDecWithPrec(5, 1),
					C:             sdkmath.LegacyNewDec(int64(9_375_000)),
					BondingTarget: sdkmath.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   sdkmath.LegacyNewDecWithPrec(20, 2).Neg(),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - inflation distribution - negative staking rewards",
			Params{
				MintDenom:              "acanto",
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdkmath.LegacyOneDec().Neg(),
					CommunityPool:  sdkmath.LegacyNewDecWithPrec(133333, 6),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation distribution - negative usage incentives",
			Params{
				MintDenom:              "acanto",
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdkmath.LegacyNewDecWithPrec(533334, 6),
					CommunityPool:  sdkmath.LegacyNewDecWithPrec(133333, 6),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation distribution - negative community pool rewards",
			Params{
				MintDenom:              "acanto",
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdkmath.LegacyNewDecWithPrec(533334, 6),
					CommunityPool:  sdkmath.LegacyOneDec().Neg(),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation distribution - total distribution ratio unequal 1",
			Params{
				MintDenom:              "acanto",
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdkmath.LegacyNewDecWithPrec(533333, 6),
					CommunityPool:  sdkmath.LegacyNewDecWithPrec(133333, 6),
				},
				EnableInflation: true,
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}
