package simulation

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
				return fmt.Sprintf("\"%s\"", sdk.NewDecWithPrec(r.Int63n(3), 3)) // 0.1%~0.3%
			},
		),
	}
}
