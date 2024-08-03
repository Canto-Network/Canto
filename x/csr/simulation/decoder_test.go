package simulation_test

import (
	"fmt"
	"testing"

	"github.com/Canto-Network/Canto/v8/x/csr/keeper"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/evmos/ethermint/tests"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/Canto-Network/Canto/v8/x/csr"
	"github.com/Canto-Network/Canto/v8/x/csr/simulation"
	"github.com/Canto-Network/Canto/v8/x/csr/types"
)

func TestCsrStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(csr.AppModuleBasic{}).Codec
	dec := simulation.NewDecodeStore(cdc)

	csr := types.CSR{
		Id:        1,
		Contracts: []string{tests.GenerateAddress().Hex()},
	}

	nftId := uint64(1)

	turnstile := tests.GenerateAddress()

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.KeyPrefixCSR, Value: cdc.MustMarshal(&csr)},
			{Key: types.KeyPrefixContract, Value: keeper.UInt64ToBytes(nftId)},
			{Key: types.KeyPrefixAddrs, Value: turnstile.Bytes()},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"CSR", fmt.Sprintf("%v\n%v", csr, csr)},
		{"NFTId", fmt.Sprintf("%v\n%v", nftId, nftId)},
		{"Turnstile", fmt.Sprintf("%v\n%v", turnstile, turnstile)},
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
