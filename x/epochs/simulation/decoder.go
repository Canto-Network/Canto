package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/Canto-Network/Canto/v8/x/epochs/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding farming type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixEpoch):
			var eA, eB types.EpochInfo
			cdc.MustUnmarshal(kvA.Value, &eA)
			cdc.MustUnmarshal(kvA.Value, &eB)
			return fmt.Sprintf("%v\n%v", eA, eB)

		default:
			panic(fmt.Sprintf("invalid epochs key prefix %X", kvA.Key[:1]))
		}
	}
}
