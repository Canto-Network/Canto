package inflation

import (
	"github.com/Canto-Network/Canto-Testnet-v2/v1/x/inflation/keeper"
	"github.com/Canto-Network/Canto-Testnet-v2/v1/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	ak types.AccountKeeper,
	_ types.StakingKeeper,
	data types.GenesisState,
) {
	// Ensure inflation module account is set on genesis
	if acc := ak.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the inflation module account has not been set")
	}

	// Set genesis state
	params := data.Params
	k.SetParams(ctx, params)

	period := data.Period
	k.SetPeriod(ctx, period)

	epochIdentifier := data.EpochIdentifier
	k.SetEpochIdentifier(ctx, epochIdentifier)

	epochsPerPeriod := data.EpochsPerPeriod
	k.SetEpochsPerPeriod(ctx, epochsPerPeriod)

	skippedEpochs := data.SkippedEpochs
	k.SetSkippedEpochs(ctx, skippedEpochs)

	//set the current inflationRate
	initInflation := sdk.NewDecWithPrec(100, 2)
	err := k.SetCurInflation(ctx, initInflation)
	if err != nil {
		panic(err)
	}

	k.GetInflationRate(ctx)

	// Calculate epoch mint provision
	epochMintProvision, err := k.CalculateEpochMintProvision(ctx)
	if err != nil {
		panic(err)
	}

	k.SetEpochMintProvision(ctx, epochMintProvision)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:          k.GetParams(ctx),
		Period:          k.GetPeriod(ctx),
		EpochIdentifier: k.GetEpochIdentifier(ctx),
		EpochsPerPeriod: k.GetEpochsPerPeriod(ctx),
		SkippedEpochs:   k.GetSkippedEpochs(ctx),
	}
}
