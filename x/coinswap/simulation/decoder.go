package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/Canto-Network/Canto/v8/x/coinswap/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding farming type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:], []byte(types.KeyPool)):
			var pA, pB types.Pool
			cdc.MustUnmarshal(kvA.Value, &pA)
			cdc.MustUnmarshal(kvB.Value, &pB)
			return fmt.Sprintf("%v\n%v", pA, pB)

		case bytes.Equal(kvA.Key[:], []byte(types.KeyNextPoolSequence)):
			var seqA, seqB uint64
			seqA = sdk.BigEndianToUint64(kvA.Value)
			seqB = sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%v\n%v", seqA, seqB)

		case bytes.Equal(kvA.Key[:], []byte(types.KeyPoolLptDenom)):
			var pA, pB types.Pool
			cdc.MustUnmarshal(kvA.Value, &pA)
			cdc.MustUnmarshal(kvB.Value, &pB)
			return fmt.Sprintf("%v\n%v", pA, pB)

		default:
			panic(fmt.Sprintf("invalid coinswap key prefix %X", kvA.Key[:1]))
		}
	}
}
