package simulation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Canto-Network/Canto/v8/x/epochs/types"
)

// DONTCOVER

// RandomizedGenState generates a random GenesisState for epochs.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesisState()

	epochs := []types.EpochInfo{
		{
			Identifier:              types.WeekEpochID,
			StartTime:               simState.GenTimestamp,
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   simState.GenTimestamp,
			EpochCountingStarted:    false,
		},
		{
			Identifier:              types.DayEpochID,
			StartTime:               simState.GenTimestamp,
			Duration:                time.Hour * 24,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   simState.GenTimestamp,
			EpochCountingStarted:    false,
		},
		{
			Identifier:              types.HourEpochID,
			StartTime:               simState.GenTimestamp,
			Duration:                time.Hour * 1,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   simState.GenTimestamp,
			EpochCountingStarted:    false,
		},
	}

	genesis.Epochs = epochs

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated epochs parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
