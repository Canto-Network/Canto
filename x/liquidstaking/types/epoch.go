package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (e *Epoch) Validate() error {
	if e.Duration != types.DefaultUnbondingTime {
		return ErrInvalidEpochDuration
	}
	// Comment the following lines checking StartTime when enable advance epoch mode.
	if !e.StartTime.Before(time.Now()) {
		return ErrInvalidEpochStartTime
	}
	return nil
}
