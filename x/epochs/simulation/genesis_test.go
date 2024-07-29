package simulation_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/Canto-Network/Canto/v8/x/epochs/simulation"
	"github.com/Canto-Network/Canto/v8/x/epochs/types"
)

func TestRandomizedGenState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(2)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var genState types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &genState)
	fmt.Println(genState.Epochs)
	require.Equal(t, []types.EpochInfo{
		{
			Identifier:              "week",
			StartTime:               time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			Duration:                604800000000000,
			CurrentEpoch:            0,
			CurrentEpochStartTime:   time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			EpochCountingStarted:    false,
			CurrentEpochStartHeight: 0,
		},
		{
			Identifier:              "day",
			StartTime:               time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			Duration:                86400000000000,
			CurrentEpoch:            0,
			CurrentEpochStartTime:   time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			EpochCountingStarted:    false,
			CurrentEpochStartHeight: 0,
		},
		{
			Identifier:              "hour",
			StartTime:               time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			Duration:                3600000000000,
			CurrentEpoch:            0,
			CurrentEpochStartTime:   time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			EpochCountingStarted:    false,
			CurrentEpochStartHeight: 0,
		},
	}, genState.Epochs)

}

// TestInvalidGenesisState tests invalid genesis states.
func TestInvalidGenesisState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)

	// all these tests will panic
	tests := []struct {
		simState module.SimulationState
		panicMsg string
	}{
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       cdc,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}
