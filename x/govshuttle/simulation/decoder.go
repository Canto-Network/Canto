package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Canto-Network/Canto/v8/x/govshuttle/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding farming type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:4], types.PortKey):
			var paA, paB common.Address
			paA = common.BytesToAddress(kvA.Value)
			paB = common.BytesToAddress(kvB.Value)
			return fmt.Sprintf("%v\n%v", paA, paB)

		default:
			panic(fmt.Sprintf("invalid govshuttle key prefix %X", kvA.Key[:1]))
		}
	}
}
