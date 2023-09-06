package types

import (
	"fmt"
	"time"
)

func (e *Epoch) Validate() error {
	if e.Duration <= 0 {
		return fmt.Errorf("duration must be positive: %d", e.Duration)
	}
	// Comment the following lines checking StartTime when enable advance epoch mode.
	if !e.StartTime.Before(time.Now()) {
		return ErrInvalidEpochStartTime
	}
	return nil
}
