package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
)

func TestParamKeyTable(t *testing.T) {
	require.IsType(t, paramtypes.KeyTable{}, ParamKeyTable())
	require.NotEmpty(t, ParamKeyTable())
}

func TestParamSetPairs(t *testing.T) {
	params := DefaultParams()
	require.NotEmpty(t, params.ParamSetPairs())
}

func TestParamsValidate(t *testing.T) {
	csrShares := sdkmath.LegacyNewDecWithPrec(50, 2)

	testCases := []struct {
		name   string
		params Params
		pass   bool
	}{
		{"Testing default parameters - pass", DefaultParams(), true},
		{
			"Testing another valid set of parameters - pass",
			NewParams(true, csrShares),
			true,
		},
		{
			"Testing disabling the CSR module - pass",
			NewParams(false, csrShares),
			true,
		},
		{
			"Testing all goes to csrShares - pass",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(1))},
			true,
		},
		{
			"Testing nothing goes to csrShares - pass",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(0))},
			true,
		},
		{
			"Testing empty parameters - fail",
			Params{},
			false,
		},
		{
			"Testing CSR shares going over 100% - fail",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(2))},
			false,
		},
		{
			"Testing CSR shares below 0 - fail",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(-1))},
			false,
		},
		{
			"Testing CSR shares below 0 - fail",
			Params{true, sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(-1))},
			false,
		},
	}
	for _, tc := range testCases {
		err := tc.params.Validate()

		if !tc.pass {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}
