package simulation

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(
			types.ModuleName, string(types.KeyFee),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomFee(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(
			types.ModuleName, string(types.KeyPoolCreationFee),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomPoolCreationFee(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(
			types.ModuleName, string(types.KeyTaxRate),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomTaxRate(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(
			types.ModuleName, string(types.KeyMaxStandardCoinPerPool),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomMaxStandardCoinPerPool(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(
			types.ModuleName, string(types.KeyMaxSwapAmount),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(generateRandomMaxSwapAmount(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
	}
}
