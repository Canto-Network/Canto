package simulation_test

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/simulation"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestParamChange(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{
			"liquidstaking/DynamicFeeRate",
			"DynamicFeeRate",
			`{"r0":"0.003951054939003790","u_soft_cap":"0.052409339630583440","u_hard_cap":"0.127604017677078046","u_optimal":"0.072579683278078640","slope1":"0.004966872261695090","slope2":"0.446589949261959746","max_fee_rate":"0.455507876462658626"}`,
			"liquidstaking",
		},
		{
			"liquidstaking/MaximumDiscountRate",
			"MaximumDiscountRate",
			`"0.057488122528113873"`,
			"liquidstaking",
		},
	}

	paramChanges := simulation.ParamChanges(r)
	require.Len(t, paramChanges, 2)

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
