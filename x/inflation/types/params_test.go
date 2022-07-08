package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	validExponentialCalculation := DefaultParams().ExponentialCalculation

	validInflationDistribution := InflationDistribution{
		StakingRewards: sdk.NewDecWithPrec(80, 2),
		CommunityPool:  sdk.NewDecWithPrec(20, 2),
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
					MinInflation:  sdk.NewDec(int64(-1)),
					MaxInflation:  sdk.NewDecWithPrec(5, 1),
					AdjustSpeed:   sdk.NewDec(int64(9_375_000)),
					BondingTarget: sdk.NewDecWithPrec(50, 2),
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
					StakingRewards: sdk.OneDec().Neg(),
					CommunityPool:  sdk.NewDecWithPrec(133333, 6),
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
					StakingRewards: sdk.NewDecWithPrec(533334, 6),
					CommunityPool:  sdk.NewDecWithPrec(133333, 6),
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
					StakingRewards: sdk.NewDecWithPrec(533334, 6),
					CommunityPool:  sdk.OneDec().Neg(),
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
					StakingRewards: sdk.NewDecWithPrec(80, 2),
					CommunityPool:  sdk.NewDecWithPrec(12, 2),
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
