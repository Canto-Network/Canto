package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/Canto-Network/Canto/v8/x/erc20/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding farming type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixTokenPair):
			var tpA, tpB types.TokenPair
			cdc.MustUnmarshal(kvA.Value, &tpA)
			cdc.MustUnmarshal(kvB.Value, &tpB)
			return fmt.Sprintf("%v\n%v", tpA, tpB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixTokenPairByERC20Address):
			var tpA, tpB types.TokenPair
			cdc.MustUnmarshal(kvA.Value, &tpA)
			cdc.MustUnmarshal(kvB.Value, &tpB)
			return fmt.Sprintf("%v\n%v", tpA, tpB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixTokenPairByDenom):
			var tpA, tpB types.TokenPair
			cdc.MustUnmarshal(kvA.Value, &tpA)
			cdc.MustUnmarshal(kvB.Value, &tpB)
			return fmt.Sprintf("%v\n%v", tpA, tpB)

		default:
			panic(fmt.Sprintf("invalid erc20 key prefix %X", kvA.Key[:1]))
		}
	}
}
