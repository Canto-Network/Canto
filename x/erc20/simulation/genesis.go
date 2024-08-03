package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Canto-Network/Canto/v8/x/erc20/types"
)

// DONTCOVER

// simulation parameter constants
const (
	enableErc20   = "enable_erc20"
	enableEVMHook = "enable_evm_hook"
)

func generateRandomBool(r *rand.Rand) bool {
	return r.Int63()%2 == 0
}

func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesisState()

	simState.AppParams.GetOrGenerate(
		enableErc20, &genesis.Params.EnableErc20, simState.Rand,
		func(r *rand.Rand) { genesis.Params.EnableErc20 = generateRandomBool(r) },
	)

	simState.AppParams.GetOrGenerate(
		enableEVMHook, &genesis.Params.EnableEVMHook, simState.Rand,
		func(r *rand.Rand) { genesis.Params.EnableEVMHook = generateRandomBool(r) },
	)

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated erc20 parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
