package simulation

import (
	"encoding/json"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"math/rand"
)

func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyDynamicFeeRate),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(genDynamicFeeRate(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
	}
}
