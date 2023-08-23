package simulation

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Canto-Network/Canto/v6/x/govshuttle/types"
)

// DONTCOVER

func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated erc20 parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
