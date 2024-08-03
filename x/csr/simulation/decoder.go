package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Canto-Network/Canto/v8/x/csr/keeper"
	"github.com/Canto-Network/Canto/v8/x/csr/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding farming type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixCSR):
			var cA, cB types.CSR
			cdc.MustUnmarshal(kvA.Value, &cA)
			cdc.MustUnmarshal(kvB.Value, &cB)
			return fmt.Sprintf("%v\n%v", cA, cB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixContract):
			var nftA, nftB uint64
			nftA = keeper.BytesToUInt64(kvA.Value)
			nftB = keeper.BytesToUInt64(kvB.Value)
			return fmt.Sprintf("%v\n%v", nftA, nftB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixAddrs):
			var tsA, tsB common.Address
			tsA = common.BytesToAddress(kvA.Value)
			tsB = common.BytesToAddress(kvB.Value)
			return fmt.Sprintf("%v\n%v", tsA, tsB)

		default:
			panic(fmt.Sprintf("invalid csr key prefix %X", kvA.Key[:1]))
		}
	}
}
