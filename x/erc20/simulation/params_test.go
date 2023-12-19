package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Canto-Network/Canto/v7/x/erc20/simulation"
)

func TestParamChanges(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	paramChanges := simulation.ParamChanges(r)
	require.Len(t, paramChanges, 2)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{"erc20/EnableErc20", "EnableErc20", "false", "erc20"},
		{"erc20/EnableEVMHook", "EnableEVMHook", "true", "erc20"},
	}

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
