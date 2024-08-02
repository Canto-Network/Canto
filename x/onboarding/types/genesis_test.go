package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestGenesisValidate(t *testing.T) {
	testCases := []struct {
		name     string
		genesis  GenesisState
		expError bool
	}{
		{
			"default genesis",
			*DefaultGenesisState(),
			false,
		},
		{
			"custom genesis",
			NewGenesisState(NewParams(true, sdkmath.NewInt(10000), []string{"channel-0"})),
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.genesis.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}
