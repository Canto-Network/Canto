package simulation_test

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/simulation"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/require"
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
		InitialStake: sdk.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var genState types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &genState)

	require.Equal(t, types.DefaultLiquidBondDenom, genState.LiquidBondDenom)
	require.Equal(t, types.DefaultR0, genState.Params.DynamicFeeRate.R0)
	require.Equal(t, types.DefaultUSoftCap, genState.Params.DynamicFeeRate.USoftCap)
	require.Equal(t, types.DefaultUHardCap, genState.Params.DynamicFeeRate.UHardCap)
	require.Equal(t, types.DefaultUOptimal, genState.Params.DynamicFeeRate.UOptimal)
	require.Equal(t, types.DefaultSlope1, genState.Params.DynamicFeeRate.Slope1)
	require.Equal(t, types.DefaultSlope2, genState.Params.DynamicFeeRate.Slope2)
	require.Equal(t, types.DefaultMaximumDiscountRate, genState.Params.MaximumDiscountRate)
	require.Equal(t, types.DefaultMaxFee, genState.Params.DynamicFeeRate.MaxFeeRate)
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
