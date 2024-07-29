package simulation_test

import (
	"fmt"
	"testing"

	"github.com/Canto-Network/Canto/v8/x/erc20"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/kv"
	testutil "github.com/evmos/ethermint/tests"

	"github.com/Canto-Network/Canto/v8/x/erc20/simulation"
	"github.com/Canto-Network/Canto/v8/x/erc20/types"
)

func TestERC20Store(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(erc20.AppModuleBasic{}).Codec
	dec := simulation.NewDecodeStore(cdc)

	tokenPair := types.NewTokenPair(testutil.GenerateAddress(), "coin", true, types.OWNER_MODULE)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.KeyPrefixTokenPair, Value: cdc.MustMarshal(&tokenPair)},
			{Key: types.KeyPrefixTokenPairByERC20Address, Value: cdc.MustMarshal(&tokenPair)},
			{Key: types.KeyPrefixTokenPairByDenom, Value: cdc.MustMarshal(&tokenPair)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"TokenPair", fmt.Sprintf("%v\n%v", tokenPair, tokenPair)},
		{"TokenPairByERC20", fmt.Sprintf("%v\n%v", tokenPair, tokenPair)},
		{"TokenPairByDenom", fmt.Sprintf("%v\n%v", tokenPair, tokenPair)},
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
