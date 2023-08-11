package simulation_test

import (
	"encoding/binary"
	"fmt"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/simulation"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDecodeLiquidStakingStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig()
	dec := simulation.NewDecodeStore(cdc.Marshaler)

	chunkA := types.NewChunk(1)
	chunkB := types.NewChunk(2)

	chunkIdBz := make([]byte, 8)
	binary.BigEndian.PutUint64(chunkIdBz, 1)

	chunkBzA := cdc.Marshaler.MustMarshal(&chunkA)
	chunkBzB := cdc.Marshaler.MustMarshal(&chunkB)

	accAddr := sdk.AccAddress("test")
	valAddr := sdk.ValAddress("test")
	feeRate := sdk.MustNewDecFromStr("0.1")
	insuranceA := types.NewInsurance(1, accAddr.String(), valAddr.String(), feeRate)
	insuranceB := types.NewInsurance(2, accAddr.String(), valAddr.String(), feeRate)

	withdrawReqA := types.NewWithdrawInsuranceRequest(1)
	withdrawReqB := types.NewWithdrawInsuranceRequest(2)

	oneCoin := sdk.NewCoin("test", sdk.NewInt(1))
	infoA := types.NewUnpairingForUnstakingChunkInfo(1, accAddr.String(), oneCoin)
	infoB := types.NewUnpairingForUnstakingChunkInfo(2, accAddr.String(), oneCoin)

	tests := []struct {
		name        string
		kvA, kvB    kv.Pair
		expectedLog string
		wantPanic   bool
	}{
		{
			"chunks",
			kv.Pair{Key: types.GetChunkKey(1), Value: chunkBzA},
			kv.Pair{Key: types.GetChunkKey(2), Value: chunkBzB},
			fmt.Sprintf("%v\n%v", chunkA, chunkB),
			false,
		},
		{
			"insurances",
			kv.Pair{Key: types.GetInsuranceKey(1), Value: cdc.Marshaler.MustMarshal(&insuranceA)},
			kv.Pair{Key: types.GetInsuranceKey(2), Value: cdc.Marshaler.MustMarshal(&insuranceB)},
			fmt.Sprintf("%v\n%v", insuranceA, insuranceB),
			false,
		},
		{
			"withdrawInsuranceRequests",
			kv.Pair{Key: types.GetWithdrawInsuranceRequestKey(1), Value: cdc.Marshaler.MustMarshal(&withdrawReqA)},
			kv.Pair{Key: types.GetWithdrawInsuranceRequestKey(2), Value: cdc.Marshaler.MustMarshal(&withdrawReqB)},
			fmt.Sprintf("%v\n%v", withdrawReqA, withdrawReqB),
			false,
		},
		{
			"unpairingForUnstakingChunkInfos",
			kv.Pair{Key: types.GetUnpairingForUnstakingChunkInfoKey(1), Value: cdc.Marshaler.MustMarshal(&infoA)},
			kv.Pair{Key: types.GetUnpairingForUnstakingChunkInfoKey(2), Value: cdc.Marshaler.MustMarshal(&infoB)},
			fmt.Sprintf("%v\n%v", infoA, infoB),
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() { dec(tt.kvA, tt.kvB) }, tt.name)
			} else {
				require.Equal(t, tt.expectedLog, dec(tt.kvA, tt.kvB), tt.name)
			}
		})
	}
}
