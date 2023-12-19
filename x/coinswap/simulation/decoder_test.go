package simulation_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Canto-Network/Canto/v7/x/coinswap/simulation"
	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

func TestCoinSwapStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig()
	dec := simulation.NewDecodeStore(cdc.Marshaler)

	pool := types.Pool{
		Id:                types.GetPoolId("denom1"),
		StandardDenom:     "denom2",
		CounterpartyDenom: "denom1",
		EscrowAddress:     types.GetReservePoolAddr("lptDenom").String(),
		LptDenom:          "lptDenom",
	}

	sequence := uint64(1)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: []byte(types.KeyPool), Value: cdc.Marshaler.MustMarshal(&pool)},
			{Key: []byte(types.KeyPoolLptDenom), Value: cdc.Marshaler.MustMarshal(&pool)},
			{Key: []byte(types.KeyNextPoolSequence), Value: sdk.Uint64ToBigEndian(sequence)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Pool", fmt.Sprintf("%v\n%v", pool, pool)},
		{"PoolLptDenom", fmt.Sprintf("%v\n%v", pool, pool)},
		{"NextPoolSequence", fmt.Sprintf("%v\n%v", sequence, sequence)},
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
