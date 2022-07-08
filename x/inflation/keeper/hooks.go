package keeper

import (
	"fmt"

	epochstypes "github.com/Canto-Network/Canto-Testnet-v2/v1/x/epochs/types"
	"github.com/Canto-Network/Canto-Testnet-v2/v1/x/inflation/types"
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeforeEpochStart: noop, We don't need to do anything here
func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) {
}

// AfterEpochEnd mints and allocates coins at the end of each epoch end
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	skippedEpochs := k.GetSkippedEpochs(ctx)

	// Skip inflation if it is disabled and increment number of skipped epochs
	if !params.EnableInflation {
		// check if the epochIdentifier is "day" before incrementing.
		if epochIdentifier != epochstypes.DayEpochID {
			return
		}
		skippedEpochs++

		k.SetSkippedEpochs(ctx, skippedEpochs)
		k.Logger(ctx).Debug(
			"skipping inflation mint and allocation",
			"height", ctx.BlockHeight(),
			"epoch-id", epochIdentifier,
			"epoch-number", epochNumber,
			"skipped-epochs", skippedEpochs,
		)
		return
	}

	expEpochID := k.GetEpochIdentifier(ctx)
	if epochIdentifier != expEpochID {
		return
	}


	k.GetInflationRate(ctx)
	newProvision, err := k.CalculateEpochMintProvision(ctx)
	if err != nil {
		panic(err)
	}
	k.SetEpochMintProvision(ctx, newProvision)

	// mint coins, update supply
	epochMintProvision, found := k.GetEpochMintProvision(ctx)

	if !found {
		panic("the epochMintProvision was not found")
	}

	mintedCoin := sdk.NewCoin(params.MintDenom, epochMintProvision.TruncateInt())
	staking, communityPool, err := k.MintAndAllocateInflation(ctx, mintedCoin)
	if err != nil {
		panic(err)
	}

	defer func() {
		if mintedCoin.Amount.IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "allocate", "total"},
				float32(mintedCoin.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
			)
		}
		if staking.AmountOf(mintedCoin.Denom).IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "allocate", "staking", "total"},
				float32(staking.AmountOf(mintedCoin.Denom).Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
			)
		}

		if communityPool.AmountOf(mintedCoin.Denom).IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "allocate", "community_pool", "total"},
				float32(communityPool.AmountOf(mintedCoin.Denom).Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
			)
		}
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
			sdk.NewAttribute(types.AttributeKeyEpochProvisions, newProvision.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
