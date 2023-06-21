package types_test

import (
	"testing"
	"time"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestGenesisValidate tests GenesisState.Validate method.
func TestGenesisValidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mutate      func(genState *types.GenesisState)
		expectedErr string
	}{
		{
			"default is valid",
			func(genState *types.GenesisState) {},
			"",
		},
		{
			"fail: invalid dynamic fee rate param",
			func(genState *types.GenesisState) {
				genState.Params.DynamicFeeRate = types.DynamicFeeRate{
					R0: sdk.OneDec().Add(
						sdk.NewDecWithPrec(1, 18),
					),
					USoftCap:   sdk.ZeroDec(),
					UHardCap:   sdk.ZeroDec(),
					UOptimal:   sdk.ZeroDec(),
					Slope1:     sdk.ZeroDec(),
					Slope2:     sdk.ZeroDec(),
					MaxFeeRate: sdk.ZeroDec(),
				}
			},
			"r0 should not be greater than 1",
		},
		{
			"fail: invalid epoch",
			func(genState *types.GenesisState) {
				genState.Epoch.StartTime = time.Now().Add(time.Hour * 24 * 30)
			},
			types.ErrInvalidEpochStartTime.Error(),
		},
		{
			"fail: chunk id > last chunk id",
			func(genState *types.GenesisState) {
				genState.LastChunkId = 1
				genState.Chunks = []types.Chunk{
					{
						Id: 2,
					},
				}
			},
			types.ErrInvalidChunkId.Error(),
		},
		{
			"fail: insurance id > last insurance id",
			func(genState *types.GenesisState) {
				genState.LastInsuranceId = 1
				genState.Insurances = []types.Insurance{
					{
						Id: 2,
					},
				}
			},
			types.ErrInvalidInsuranceId.Error(),
		},
		{
			"fail: unpairingForUnstakingChunkInfo exist for non-existing chunk",
			func(genState *types.GenesisState) {
				genState.UnpairingForUnstakingChunkInfos = []types.UnpairingForUnstakingChunkInfo{
					{
						ChunkId: 1,
					},
				}
			},
			types.ErrNotFoundUnpairingForUnstakingChunkInfoChunkId.Error(),
		},
		{
			"fail: withdrawInsuranceRequest exist for non-existing insurance",
			func(genState *types.GenesisState) {
				genState.WithdrawInsuranceRequests = []types.WithdrawInsuranceRequest{
					{
						InsuranceId: 1,
					},
				}
			},
			types.ErrNotFoundWithdrawInsuranceRequestInsuranceId.Error(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesisState()
			tc.mutate(genState)
			err := genState.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedErr)
			}
		})
	}
}
