package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Canto-Network/Canto/v7/x/coinswap/simulation"
)

func TestParamChanges(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	paramChanges := simulation.ParamChanges(r)
	require.Len(t, paramChanges, 5)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{"coinswap/Fee", "Fee", "\"0.004000000000000000\"", "coinswap"},
		{"coinswap/PoolCreationFee", "PoolCreationFee", `{"denom":"stake","amount":"58514"}`, "coinswap"},
		{"coinswap/TaxRate", "TaxRate", "\"0.003000000000000000\"", "coinswap"},
		{"coinswap/MaxStandardCoinPerPool", "MaxStandardCoinPerPool", "\"2506000000000000000000\"", "coinswap"},
		{"coinswap/MaxSwapAmount", "MaxSwapAmount", `[{"denom":"ibc/17CD484EE7D9723B847D95015FA3EBD1572FD13BC84FB838F55B18A57450F25B","amount":"27000000"},{"denom":"ibc/4F6A2DEFEA52CD8D90966ADCB2BD0593D3993AB0DF7F6AEB3EFD6167D79237B0","amount":"35000000"},{"denom":"ibc/DC186CA7A8C009B43774EBDC825C935CABA9743504CE6037507E6E5CCE12858A","amount":"980000000000000000"}]`, "coinswap"},
	}

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
