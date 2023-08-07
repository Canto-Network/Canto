package simulation_test

import (
	"fmt"
	"testing"

	"github.com/evmos/ethermint/tests"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/Canto-Network/Canto/v6/x/govshuttle/simulation"
	"github.com/Canto-Network/Canto/v6/x/govshuttle/types"
)

func TestGovShuttleStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig()
	dec := simulation.NewDecodeStore(cdc.Marshaler)

	portAddress := tests.GenerateAddress()

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.PortKey, Value: portAddress.Bytes()},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"PortAddress", fmt.Sprintf("%v\n%v", portAddress, portAddress)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
