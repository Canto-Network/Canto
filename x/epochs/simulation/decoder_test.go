package simulation_test

import (
	"fmt"
	"testing"
	"time"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/Canto-Network/Canto/v8/x/epochs"
	"github.com/Canto-Network/Canto/v8/x/epochs/simulation"
	"github.com/Canto-Network/Canto/v8/x/epochs/types"
)

func TestEpochsStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(epochs.AppModuleBasic{}).Codec
	dec := simulation.NewDecodeStore(cdc)

	epoch := types.EpochInfo{
		Identifier:              types.DayEpochID,
		StartTime:               time.Time{},
		Duration:                time.Hour * 24,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.KeyPrefixEpoch, Value: cdc.MustMarshal(&epoch)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Epoch", fmt.Sprintf("%v\n%v", epoch, epoch)},
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
