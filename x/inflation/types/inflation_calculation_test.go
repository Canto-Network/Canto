package types

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type InflationTestSuite struct {
	suite.Suite
}

func TestInflationSuite(t *testing.T) {
	suite.Run(t, new(InflationTestSuite))
}

func (suite *InflationTestSuite) TestCalculateEpochMintProvision() {
	bondingParams := DefaultParams()
	bondingParams.ExponentialCalculation.MaxVariance = sdk.NewDecWithPrec(0, 2)
	epochsPerPeriod := int64(30)

	testCases := []struct {
		name              string
		params            Params
		period            uint64
		bondedRatio       sdk.Dec
		expEpochProvision sdk.Dec
		expPass           bool
	}{
		{
			"pass - initial period",
			DefaultParams(),
			uint64(0),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("543478266666666666666667.000000000000000000"),
			true,
		},
		{
			"pass - period 1",
			DefaultParams(),
			uint64(1),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("353260873333333333333333.000000000000000000"),
			true,
		},
		{
			"pass - period 2",
			DefaultParams(),
			uint64(2),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("229619567666666666666667.000000000000000000"),
			true,
		},
		{
			"pass - period 3",
			DefaultParams(),
			uint64(3),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("149252718983333333333333.000000000000000000"),
			true,
		},
		{
			"pass - period 20",
			DefaultParams(),
			uint64(20),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("98502967552518961527.000000000000000000"),
			true,
		},
		{
			"pass - period 21",
			DefaultParams(),
			uint64(21),
			sdk.OneDec(),
			sdk.MustNewDecFromStr("64026928909137053253.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - initial period",
			bondingParams,
			uint64(0),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("543478266666666666666667.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 1",
			bondingParams,
			uint64(1),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("353260873333333333333333.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 2",
			bondingParams,
			uint64(2),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("229619567666666666666667.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 3",
			bondingParams,
			uint64(3),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("149252718983333333333333.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 20",
			bondingParams,
			uint64(20),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("98502967552518961527.000000000000000000"),
			true,
		},
		{
			"pass - 0 percent bonding - period 21",
			bondingParams,
			uint64(21),
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("64026928909137053253.000000000000000000"),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			epochMintProvisions := CalculateEpochMintProvision(
				tc.params,
				tc.period,
				epochsPerPeriod,
				tc.bondedRatio,
			)

			suite.Require().Equal(tc.expEpochProvision, epochMintProvisions)
		})
	}
}
