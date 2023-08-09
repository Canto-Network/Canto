package simulation

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v6/x/erc20/types"
)

func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyEnableErc20),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomBool(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyEnableEVMHook),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomBool(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
	}
}
