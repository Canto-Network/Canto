package keeper

import (
	"context"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v7/x/epochs/types"
)

// BeginBlocker of epochs module
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := k.Logger(sdkCtx)
	k.IterateEpochInfo(sdkCtx, func(_ int64, epochInfo types.EpochInfo) (stop bool) {
		// Has it not started, and is the block time > initial epoch start time
		shouldInitialEpochStart := !epochInfo.EpochCountingStarted && !epochInfo.StartTime.After(sdkCtx.BlockTime())

		epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
		shouldEpochEnd := sdkCtx.BlockTime().After(epochEndTime) && !shouldInitialEpochStart && !epochInfo.StartTime.After(sdkCtx.BlockTime())

		epochInfo.CurrentEpochStartHeight = sdkCtx.BlockHeight()

		switch {
		case shouldInitialEpochStart:
			epochInfo.StartInitialEpoch()

			logger.Info("starting epoch", "identifier", epochInfo.Identifier)
		case shouldEpochEnd:
			epochInfo.EndEpoch()

			logger.Info("ending epoch", "identifier", epochInfo.Identifier)

			sdkCtx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEpochEnd,
					sdk.NewAttribute(types.AttributeEpochNumber, strconv.FormatInt(epochInfo.CurrentEpoch, 10)),
				),
			)
			k.AfterEpochEnd(sdkCtx, epochInfo.Identifier, epochInfo.CurrentEpoch)
		default:
			// continue
			return false
		}

		k.SetEpochInfo(sdkCtx, epochInfo)

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEpochStart,
				sdk.NewAttribute(types.AttributeEpochNumber, strconv.FormatInt(epochInfo.CurrentEpoch, 10)),
				sdk.NewAttribute(types.AttributeEpochStartTime, strconv.FormatInt(epochInfo.CurrentEpochStartTime.Unix(), 10)),
			),
		)

		k.BeforeEpochStart(sdkCtx, epochInfo.Identifier, epochInfo.CurrentEpoch)

		return false
	})
	return nil
}
