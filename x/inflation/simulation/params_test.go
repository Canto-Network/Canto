package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Canto-Network/Canto/v6/x/inflation/simulation"
)

func TestParamChanges(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	paramChanges := simulation.ParamChanges(r)
	require.Len(t, paramChanges, 4)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{"inflation/ParamStoreKeyMintDenom", "ParamStoreKeyMintDenom", "\"stake\"", "inflation"},
		{"inflation/ParamStoreKeyExponentialCalculation", "ParamStoreKeyExponentialCalculation", `{"a":"9793274.000000000000000000","r":"0.140000000000000000","c":"0.000000000000000000","bonding_target":"0.530000000000000000","max_variance":"0.000000000000000000"}`, "inflation"},
		{"inflation/ParamStoreKeyInflationDistribution", "ParamStoreKeyInflationDistribution", `{"staking_rewards":"6702506.000000000000000000","community_pool":"9387515.000000000000000000"}`, "inflation"},
		{"inflation/ParamStoreKeyEnableInflation", "ParamStoreKeyEnableInflation", "false", "inflation"},
	}

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
