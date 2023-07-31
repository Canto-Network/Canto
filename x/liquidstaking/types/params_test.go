package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, ParamKeyTable())

	params := DefaultParams()
	paramStr := `dynamicfeerate:
  r0: "0.000000000000000000"
  usoftcap: "0.050000000000000000"
  uhardcap: "0.100000000000000000"
  uoptimal: "0.090000000000000000"
  slope1: "0.100000000000000000"
  slope2: "0.400000000000000000"
  maxfeerate: "0.500000000000000000"
maximumdiscountrate: "0.030000000000000000"
`

	require.Equal(t, paramStr, params.String())
}

// TestValidateParamsBasic tests basic validation
// of each params fields in DynamicFeeRate.
func TestValidateParamsBasic(t *testing.T) {
	negativeDec := sdk.NewDecFromInt(sdk.NewInt(-1))
	biggerThanOneDec := sdk.OneDec().Add(
		sdk.NewDecWithPrec(1, 18),
	)
	for _, tc := range []struct {
		name       string
		setupParam func(*Params)
		errStr     string
	}{
		{
			"validate default params",
			func(params *Params) {},
			"",
		},
		{
			"invalid r0 - negative",
			func(params *Params) {
				params.DynamicFeeRate.R0 = negativeDec
			},
			"r0 should not be negative",
		},
		{
			"invalid r0 - bigger than 1",
			func(params *Params) {
				params.DynamicFeeRate.R0 = biggerThanOneDec
			},
			"r0 should not be greater than 1",
		},
		{
			"invalid uSoftCap - bigger than 1",
			func(params *Params) {
				params.DynamicFeeRate.USoftCap = biggerThanOneDec
			},
			"uSoftCap should not be greater than 1",
		},
		{
			"invalid uSoftCap - negative",
			func(params *Params) {
				params.DynamicFeeRate.USoftCap = negativeDec
			},
			"uSoftCap should not be negative",
		},
		{
			"invalid uHardCap - bigger than 1",
			func(params *Params) {
				params.DynamicFeeRate.UHardCap = sdk.OneDec().Add(
					sdk.NewDecWithPrec(1, 18),
				)
			},
			"uHardCap should not be greater than 1",
		},
		{
			"invalid uHardCap - negative",
			func(params *Params) {
				params.DynamicFeeRate.UHardCap = negativeDec
			},
			"uHardCap should not be negative",
		},
		{
			"invalid uOptimal - bigger than 1",
			func(params *Params) {
				params.DynamicFeeRate.UOptimal = sdk.OneDec().Add(
					sdk.NewDecWithPrec(1, 18),
				)
			},
			"uOptimal should not be greater than 1",
		},
		{
			"invalid uOptimal - negative",
			func(params *Params) {
				params.DynamicFeeRate.UOptimal = negativeDec
			},
			"uOptimal should not be negative",
		},
		{
			"invalid slope1 - negative",
			func(params *Params) {
				params.DynamicFeeRate.Slope1 = negativeDec
			},
			"slope1 should not be negative",
		},
		{
			"invalid slope2 - negative",
			func(params *Params) {
				params.DynamicFeeRate.Slope2 = negativeDec
			},
			"slope2 should not be negative",
		},
		{
			"invalid maxFeeRate - bigger than 1",
			func(params *Params) {
				params.DynamicFeeRate.MaxFeeRate = biggerThanOneDec
			},
			"maxFeeRate should not be greater than 1",
		},
		{
			"invalid maxFeeRate - negative",
			func(params *Params) {
				params.DynamicFeeRate.MaxFeeRate = negativeDec
			},
			"maxFeeRate should not be negative",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := DefaultParams()
			tc.setupParam(&params)
			err := params.Validate()
			if tc.errStr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.errStr)
			}
		})
	}
}

// TestValidParamsDynamicFeeModel tests checks whether each params
// follows the rules of DynamicFeeRate model.
func TestValidParamsDynamicFeeModel(t *testing.T) {
	for _, tc := range []struct {
		name       string
		setupParam func(*Params)
		errStr     string
	}{
		{
			"uSoftCap > uOptimal",
			func(params *Params) {
				params.DynamicFeeRate.USoftCap = sdk.MustNewDecFromStr("0.1")
				params.DynamicFeeRate.UOptimal = sdk.MustNewDecFromStr("0.09")
			},
			"uSoftCap should be less than uOptimal",
		},
		{
			"uSoftCap == uOptimal",
			func(params *Params) {
				params.DynamicFeeRate.USoftCap = sdk.MustNewDecFromStr("0.09")
				params.DynamicFeeRate.UOptimal = sdk.MustNewDecFromStr("0.09")
			},
			"uSoftCap should be less than uOptimal",
		},
		{
			"uOptimal > uHardCap",
			func(params *Params) {
				params.DynamicFeeRate.UOptimal = sdk.MustNewDecFromStr("0.09")
				params.DynamicFeeRate.UHardCap = sdk.MustNewDecFromStr("0.08")
			},
			"uOptimal should be less than uHardCap",
		},
		{
			"uOptimal == uHardCap",
			func(params *Params) {
				params.DynamicFeeRate.UOptimal = sdk.MustNewDecFromStr("0.09")
				params.DynamicFeeRate.UHardCap = sdk.MustNewDecFromStr("0.09")
			},
			"uOptimal should be less than uHardCap",
		},
		{
			"r0 + slope1 + slope2 > maxFeeRate",
			func(params *Params) {
				params.DynamicFeeRate.R0 = sdk.MustNewDecFromStr("0.01")
				params.DynamicFeeRate.Slope1 = sdk.MustNewDecFromStr("0.2")
				params.DynamicFeeRate.Slope2 = sdk.MustNewDecFromStr("0.4")
				params.DynamicFeeRate.MaxFeeRate = sdk.MustNewDecFromStr("0.5")
			},
			"r0 + slope1 + slope2 should not exceeds maxFeeRate",
		},
		{
			"OK: uSoftCap < uOptimal",
			func(params *Params) {
				params.DynamicFeeRate.USoftCap = sdk.MustNewDecFromStr("0.05")
				params.DynamicFeeRate.UOptimal = sdk.MustNewDecFromStr("0.09")
			},
			"",
		},
		{
			"OK: uOptimal < uHardCap",
			func(params *Params) {
				params.DynamicFeeRate.UOptimal = sdk.MustNewDecFromStr("0.09")
				params.DynamicFeeRate.UHardCap = sdk.MustNewDecFromStr("0.1")
			},
			"",
		},

		{
			"OK: r0 + slope1 + slope2 == maxFeeRate",
			func(params *Params) {
				params.DynamicFeeRate.R0 = sdk.MustNewDecFromStr("0.01")
				params.DynamicFeeRate.Slope1 = sdk.MustNewDecFromStr("0.2")
				params.DynamicFeeRate.Slope2 = sdk.MustNewDecFromStr("0.29")
				params.DynamicFeeRate.MaxFeeRate = sdk.MustNewDecFromStr("0.5")
			},
			"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := DefaultParams()
			tc.setupParam(&params)
			err := params.Validate()
			if tc.errStr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.errStr)
			}
		})
	}
}
