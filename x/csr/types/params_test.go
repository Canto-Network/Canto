package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	csrShares := sdk.NewDecWithPrec(50, 2)
	addressDerivationCostCreate := uint64(50)

	testCases := []struct {
		name   string
		params Params
		pass   bool
	}{
		{"Testing default parameters - pass", DefaultParams(), true},
		{
			"Testing another valid set of parameters - pass",
			NewParams(true, csrShares, addressDerivationCostCreate),
			true,
		},
		{
			"Testing disabling the CSR module - pass",
			NewParams(false, csrShares, addressDerivationCostCreate),
			true,
		},
		{
			"Testing all goes to csrShares - pass",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(1)), addressDerivationCostCreate},
			true,
		},
		{
			"Testing all nothing goes to csrShares - pass",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(0)), addressDerivationCostCreate},
			true,
		},
		{
			"Testing empty parameters - fail",
			Params{},
			false,
		},
		{
			"Testing CSR shares going over 100% - fail",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(2)), addressDerivationCostCreate},
			false,
		},
		{
			"Testing CSR shares below 0 - fail",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(-1)), addressDerivationCostCreate},
			false,
		},
		{
			"Testing CSR shares below 0 - fail",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(-1)), addressDerivationCostCreate},
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
