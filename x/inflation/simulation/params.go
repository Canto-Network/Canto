package simulation

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v6/x/inflation/types"
)

func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyMintDenom),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateMintDenom(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyExponentialCalculation),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateExponentialCalculation(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyInflationDistribution),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateInflationDistribution(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyEnableInflation),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateEnableInflation(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
	}
}
