package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (e *Epoch) Validate() error {
	if e.Duration != types.DefaultUnbondingTime {
		return ErrInvalidEpochDuration
	}
	if !e.StartTime.Before(time.Now()) {
		return ErrInvalidEpochStartTime
	}
	return nil
}
