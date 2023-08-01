package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/Canto-Network/Canto/v6/x/csr/types"
)

// DONTCOVER

// simulation parameter constants
const (
	enableCsr = "enable_csr"
	csrShares = "csr_shares"
)

func generateRandomBool(r *rand.Rand) bool {
	return r.Int63()%2 == 0
}

func generateRandomCsrShares(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 100)), 2)
}

// RandomizedGenState generates a random GenesisState for CSR.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, enableCsr, &genesis.Params.EnableCsr, simState.Rand,
		func(r *rand.Rand) { genesis.Params.EnableCsr = generateRandomBool(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, csrShares, &genesis.Params.CsrShares, simState.Rand,
		func(r *rand.Rand) { genesis.Params.CsrShares = generateRandomCsrShares(r) },
	)

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated inflation parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
