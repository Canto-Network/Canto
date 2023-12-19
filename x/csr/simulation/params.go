package simulation

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v7/x/csr/types"
)

func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyEnableCSR),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomBool(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyCSRShares),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomCsrShares(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
	}
}
