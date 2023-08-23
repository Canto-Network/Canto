package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/Canto-Network/Canto/v7/x/coinswap/simulation"
	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
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

	require.Equal(t, sdk.NewDecWithPrec(4, 3), genState.Params.Fee)
	require.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, 163511), genState.Params.PoolCreationFee)
	require.Equal(t, sdk.NewDecWithPrec(6, 3), genState.Params.TaxRate)
	require.Equal(t, sdk.NewIntWithDecimal(3310, 18), genState.Params.MaxStandardCoinPerPool)
	require.Equal(t, sdk.NewCoins(
		sdk.NewCoin(types.UsdcIBCDenom, sdk.NewIntWithDecimal(70, 6)),
		sdk.NewCoin(types.UsdtIBCDenom, sdk.NewIntWithDecimal(52, 6)),
		sdk.NewCoin(types.EthIBCDenom, sdk.NewIntWithDecimal(65, 16)),
	), genState.Params.MaxSwapAmount)

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
