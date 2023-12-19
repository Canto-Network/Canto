package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/Canto-Network/Canto/v7/x/inflation/simulation"
	"github.com/Canto-Network/Canto/v7/x/inflation/types"
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
		InitialStake: 1000,
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var genState types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &genState)

	require.Equal(t, "stake", genState.Params.MintDenom)
	require.Equal(t, types.ExponentialCalculation{
		A:             sdk.NewDec(2712964),
		R:             sdk.NewDecWithPrec(11, 2),
		C:             sdk.ZeroDec(),
		BondingTarget: sdk.NewDecWithPrec(94, 2),
		MaxVariance:   sdk.ZeroDec(),
	}, genState.Params.ExponentialCalculation)
	require.Equal(t, types.InflationDistribution{
		StakingRewards: sdk.NewDecWithPrec(1, 1),
		CommunityPool:  sdk.NewDecWithPrec(9, 1),
	}, genState.Params.InflationDistribution)
	require.Equal(t, false, genState.Params.EnableInflation)
	require.Equal(t, uint64(1654145), genState.Period)
	require.Equal(t, "day", genState.EpochIdentifier)
	require.Equal(t, int64(6634432), genState.EpochsPerPeriod)
	require.Equal(t, uint64(5142676), genState.SkippedEpochs)

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
