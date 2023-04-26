package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
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
			NewGenesisState(NewParams(true, sdk.NewInt(10000), []string{"channel-0"})),
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
