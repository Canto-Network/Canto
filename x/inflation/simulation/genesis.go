package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/Canto-Network/Canto/v8/x/inflation/types"
)

// DONTCOVER

// simulation parameter constants
const (
	mintDenom              = "mint_denom"
	exponentialCalculation = "exponential_calculation"
	inflationDistribution  = "inflation_distribution"
	enableInflation        = "enable_inflation"
	period                 = "period"
	epochIdentifier        = "epoch_identifier"
	epochsPerPeriod        = "epochs_per_period"
	skippedEpochs          = "skipped_epochs"
)

func generateRandomBool(r *rand.Rand) bool {
	return r.Int63()%2 == 0
}

func generateMintDenom(r *rand.Rand) string {
	return sdk.DefaultBondDenom
}

func generateExponentialCalculation(r *rand.Rand) types.ExponentialCalculation {
	return types.ExponentialCalculation{
		A:             sdkmath.LegacyNewDec(int64(simtypes.RandIntBetween(r, 0, 10000000))),
		R:             sdkmath.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 100)), 2),
		C:             sdkmath.LegacyZeroDec(),
		BondingTarget: sdkmath.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 100)), 2),
		MaxVariance:   sdkmath.LegacyZeroDec(),
	}
}

func generateInflationDistribution(r *rand.Rand) types.InflationDistribution {

	stakingRewards := sdkmath.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 100)), 2)
	communityPool := sdkmath.LegacyNewDec(1).Sub(stakingRewards)

	return types.InflationDistribution{
		StakingRewards: stakingRewards,
		CommunityPool:  communityPool,
	}
}

func generateEnableInflation(r *rand.Rand) bool {
	return generateRandomBool(r)
}

func generatePeriod(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 0, 10000000))
}

func generateEpochIdentifier(r *rand.Rand) string {
	return "day"
}

func generateEpochsPerPeriod(r *rand.Rand) int64 {
	return int64(simtypes.RandIntBetween(r, 0, 10000000))
}

func generateSkippedEpochs(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 0, 10000000))
}

// RandomizedGenState generates a random GenesisState for inflation.

func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesisState()

	simState.AppParams.GetOrGenerate(
		mintDenom, &genesis.Params.MintDenom, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MintDenom = generateMintDenom(r) },
	)

	simState.AppParams.GetOrGenerate(
		exponentialCalculation, &genesis.Params.ExponentialCalculation, simState.Rand,
		func(r *rand.Rand) { genesis.Params.ExponentialCalculation = generateExponentialCalculation(r) },
	)

	simState.AppParams.GetOrGenerate(
		inflationDistribution, &genesis.Params.InflationDistribution, simState.Rand,
		func(r *rand.Rand) { genesis.Params.InflationDistribution = generateInflationDistribution(r) },
	)

	simState.AppParams.GetOrGenerate(
		enableInflation, &genesis.Params.EnableInflation, simState.Rand,
		func(r *rand.Rand) { genesis.Params.EnableInflation = generateEnableInflation(r) },
	)

	simState.AppParams.GetOrGenerate(
		period, &genesis.Period, simState.Rand,
		func(r *rand.Rand) { genesis.Period = generatePeriod(r) },
	)

	simState.AppParams.GetOrGenerate(
		epochIdentifier, &genesis.EpochIdentifier, simState.Rand,
		func(r *rand.Rand) { genesis.EpochIdentifier = generateEpochIdentifier(r) },
	)

	simState.AppParams.GetOrGenerate(
		epochsPerPeriod, &genesis.EpochsPerPeriod, simState.Rand,
		func(r *rand.Rand) { genesis.EpochsPerPeriod = generateEpochsPerPeriod(r) },
	)

	simState.AppParams.GetOrGenerate(
		skippedEpochs, &genesis.SkippedEpochs, simState.Rand,
		func(r *rand.Rand) { genesis.SkippedEpochs = generateSkippedEpochs(r) },
	)

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated inflation parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
