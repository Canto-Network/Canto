package simulation

import (
	"bytes"
	"fmt"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding distribution type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixLastChunkId),
			bytes.Equal(kvA.Key[:1], types.KeyPrefixLastInsuranceId):
			lastIdA := sdk.BigEndianToUint64(kvA.Value)
			lastIdB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%v\n%v", lastIdA, lastIdB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixChunk):
			var chunkA, chunkB types.Chunk
			cdc.MustUnmarshal(kvA.Value, &chunkA)
			cdc.MustUnmarshal(kvB.Value, &chunkB)
			return fmt.Sprintf("%v\n%v", chunkA, chunkB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixInsurance):
			var insuranceA, insuranceB types.Insurance
			cdc.MustUnmarshal(kvA.Value, &insuranceA)
			cdc.MustUnmarshal(kvB.Value, &insuranceB)
			return fmt.Sprintf("%v\n%v", insuranceA, insuranceB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixWithdrawInsuranceRequest):
			var withdrawInsuranceRequestA, withdrawInsuranceRequestB types.WithdrawInsuranceRequest
			cdc.MustUnmarshal(kvA.Value, &withdrawInsuranceRequestA)
			cdc.MustUnmarshal(kvB.Value, &withdrawInsuranceRequestB)
			return fmt.Sprintf("%v\n%v", withdrawInsuranceRequestA, withdrawInsuranceRequestB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixUnpairingForUnstakingChunkInfo):
			var unpairingForUnstakingChunkInfoA, unpairingForUnstakingChunkInfoB types.UnpairingForUnstakingChunkInfo
			cdc.MustUnmarshal(kvA.Value, &unpairingForUnstakingChunkInfoA)
			cdc.MustUnmarshal(kvB.Value, &unpairingForUnstakingChunkInfoB)
			return fmt.Sprintf("%v\n%v", unpairingForUnstakingChunkInfoA, unpairingForUnstakingChunkInfoB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixRedelegationInfo):
			var redelegationInfoA, redelegationInfoB types.RedelegationInfo
			cdc.MustUnmarshal(kvA.Value, &redelegationInfoA)
			cdc.MustUnmarshal(kvB.Value, &redelegationInfoB)
			return fmt.Sprintf("%v\n%v", redelegationInfoA, redelegationInfoB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixEpoch):
			var epochA, epochB types.Epoch
			cdc.MustUnmarshal(kvA.Value, &epochA)
			cdc.MustUnmarshal(kvB.Value, &epochB)
			return fmt.Sprintf("%v\n%v", epochA, epochB)

		case bytes.Equal(kvA.Key[:1], types.KeyPrefixLiquidBondDenom):
			return fmt.Sprintf("%v\n%v", kvA.Value, kvB.Value)

		default:
			panic(fmt.Sprintf("invalid liquidstaking key prefix %X", kvA.Key[:1]))
		}
	}
}
