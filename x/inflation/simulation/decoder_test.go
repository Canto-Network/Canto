package simulation_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/Canto-Network/Canto/v7/x/inflation/simulation"
	"github.com/Canto-Network/Canto/v7/x/inflation/types"
)

func TestInflationStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig()
	dec := simulation.NewDecodeStore(cdc.Marshaler)

	period := uint64(1)
	epochMintProvision := sdk.NewDec(2)
	epochIdentifier := "epochIdentifier"
	epochPerPeriod := uint64(3)
	skippedEpoch := uint64(4)

	marshaled, _ := epochMintProvision.Marshal()

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.KeyPrefixPeriod, Value: sdk.Uint64ToBigEndian(period)},
			{Key: types.KeyPrefixEpochMintProvision, Value: marshaled},
			{Key: types.KeyPrefixEpochIdentifier, Value: []byte(epochIdentifier)},
			{Key: types.KeyPrefixEpochsPerPeriod, Value: sdk.Uint64ToBigEndian(epochPerPeriod)},
			{Key: types.KeyPrefixSkippedEpochs, Value: sdk.Uint64ToBigEndian(skippedEpoch)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Period", fmt.Sprintf("%v\n%v", period, period)},
		{"EpochMintProvision", fmt.Sprintf("%v\n%v", epochMintProvision, epochMintProvision)},
		{"EpochIdentifier", fmt.Sprintf("%v\n%v", epochIdentifier, epochIdentifier)},
		{"EpochsPerPeriod", fmt.Sprintf("%v\n%v", epochPerPeriod, epochPerPeriod)},
		{"SkippedEpochs", fmt.Sprintf("%v\n%v", skippedEpoch, skippedEpoch)},
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
