package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// simulation parameter constants
const (
	fee                    = "fee"
	poolCreationFee        = "pool_creation_fee"
	taxRate                = "tax_rate"
	maxStandardCoinPerPool = "max_standard_coin_per_pool"
	maxSwapAmount          = "max_swap_amount"
)

func generateRandomFee(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 10)), 3)
}

func generateRandomPoolCreationFee(r *rand.Rand) sdk.Coin {
	return sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simtypes.RandIntBetween(r, 0, 1000000)))
}

func generateRandomTaxRate(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 10)), 3)
}

func generateRandomMaxStandardCoinPerPool(r *rand.Rand) sdk.Int {
	return sdk.NewIntWithDecimal(int64(simtypes.RandIntBetween(r, 0, 10000)), 18)
}

func generateRandomMaxSwapAmount(r *rand.Rand) sdk.Coins {
	return sdk.NewCoins(
		sdk.NewCoin(types.UsdcIBCDenom, sdk.NewIntWithDecimal(int64(simtypes.RandIntBetween(r, 1, 100)), 6)),
		sdk.NewCoin(types.UsdtIBCDenom, sdk.NewIntWithDecimal(int64(simtypes.RandIntBetween(r, 1, 100)), 6)),
		sdk.NewCoin(types.EthIBCDenom, sdk.NewIntWithDecimal(int64(simtypes.RandIntBetween(r, 1, 100)), 16)),
	)
}

// RandomizedGenState generates a random GenesisState for coinswap
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesisState()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, fee, &genesis.Params.Fee, simState.Rand,
		func(r *rand.Rand) { genesis.Params.Fee = generateRandomFee(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, poolCreationFee, &genesis.Params.PoolCreationFee, simState.Rand,
		func(r *rand.Rand) { genesis.Params.PoolCreationFee = generateRandomPoolCreationFee(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, taxRate, &genesis.Params.TaxRate, simState.Rand,
		func(r *rand.Rand) { genesis.Params.TaxRate = generateRandomTaxRate(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, maxStandardCoinPerPool, &genesis.Params.MaxStandardCoinPerPool, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxStandardCoinPerPool = generateRandomMaxStandardCoinPerPool(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, maxSwapAmount, &genesis.Params.MaxSwapAmount, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxSwapAmount = generateRandomMaxSwapAmount(r) },
	)

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated coinswap parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)

}
