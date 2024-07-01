package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestParamsMisc(t *testing.T) {
	params := DefaultParams()
	require.NotEmpty(t, params.ParamSetPairs())
	kt := ParamKeyTable()
	require.NotEmpty(t, kt)
}

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"default params",
			DefaultParams(),
			false,
		},
		{
			"custom params",
			NewParams(true, sdkmath.NewInt(10000), []string{"channel-0"}),
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestValidate(t *testing.T) {
	require.Error(t, validateBool(""))
	require.NoError(t, validateBool(true))
}
