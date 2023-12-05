package simulation

import (
	"fmt"
	"math/rand"

	sdkmath "cosmossdk.io/math"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		simulation.NewSimLegacyParamChange(
			types.ModuleName, string(types.KeyFee),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", sdkmath.LegacyNewDecWithPrec(r.Int63n(3), 3)) // 0.1%~0.3%
			},
		),
	}
}
