package simulation

import (
	"bytes"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/Canto-Network/Canto/v8/x/inflation/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding farming type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixPeriod):
			var pA, pB uint64
			pA = sdk.BigEndianToUint64(kvA.Value)
			pB = sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%v\n%v", pA, pB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixEpochMintProvision):
			var empA, empB sdkmath.LegacyDec
			empA.Unmarshal(kvA.Value)
			empB.Unmarshal(kvB.Value)
			return fmt.Sprintf("%v\n%v", empA, empB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixEpochIdentifier):
			var eiA, eiB string
			eiA = string(kvA.Value)
			eiB = string(kvB.Value)
			return fmt.Sprintf("%v\n%v", eiA, eiB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixEpochsPerPeriod):
			var eppA, eppB uint64
			eppA = sdk.BigEndianToUint64(kvA.Value)
			eppB = sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%v\n%v", eppA, eppB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixSkippedEpochs):
			var seA, seB uint64
			seA = sdk.BigEndianToUint64(kvA.Value)
			seB = sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%v\n%v", seA, seB)

		default:
			panic(fmt.Sprintf("invalid farming key prefix %X", kvA.Key[:1]))
		}
	}
}
