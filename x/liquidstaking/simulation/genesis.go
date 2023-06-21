package simulation

import (
	"encoding/json"
	"fmt"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"math/rand"
)

// Simulation parameter constants
const (
	dynamicFeeRate = "dynamic_fee_rate"
)

func GenDynamicFeeRate(r *rand.Rand) types.DynamicFeeRate {
	maxFeeRate := types.RandomDec(r, sdk.MustNewDecFromStr("0.5"), sdk.MustNewDecFromStr("0.8"))

	r0 := types.RandomDec(r, sdk.ZeroDec(), sdk.MustNewDecFromStr("0.01"))
	slope1 := types.RandomDec(r, sdk.ZeroDec(), sdk.MustNewDecFromStr("0.3"))
	slope2 := maxFeeRate.Sub(slope1).Sub(r0)

	uSoftCap := types.RandomDec(r, sdk.ZeroDec(), sdk.MustNewDecFromStr("0.1"))
	uOptimal := types.RandomDec(r, uSoftCap.Add(sdk.OneDec()), sdk.MustNewDecFromStr("0.15"))
	uHardCap := types.RandomDec(r, uOptimal.Add(sdk.OneDec()), types.SecurityCap)

	return types.DynamicFeeRate{
		R0:         r0,
		USoftCap:   uSoftCap,
		UHardCap:   uHardCap,
		UOptimal:   uOptimal,
		Slope1:     slope1,
		Slope2:     slope2,
		MaxFeeRate: maxFeeRate,
	}
}

func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesisState()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, dynamicFeeRate, &genesis.Params.DynamicFeeRate, simState.Rand,
		func(r *rand.Rand) { genesis.Params.DynamicFeeRate = GenDynamicFeeRate(r) },
	)

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated liquidstaking parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
