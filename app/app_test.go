package app

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/Canto-Network/Canto/v7/types"
)

func TestCantoExport(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewCanto(
		log.NewLogger(os.Stdout),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		false,
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(types.MainnetChainID+"-1"),
	)

	genesisState := NewDefaultGenesisState()
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		&abci.RequestInitChain{
			ChainId:         types.MainnetChainID + "-1",
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
	})
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewCanto(
		log.NewLogger(os.Stdout),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		false,
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(types.MainnetChainID+"-1"),
	)
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
