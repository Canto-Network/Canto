package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Canto-Network/Canto/v6/x/csr/simulation"
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
		{"csr/EnableCSR", "EnableCSR", "false", "csr"},
		{"csr/CSRShares", "CSRShares", `"0.140000000000000000"`, "csr"},
	}

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
