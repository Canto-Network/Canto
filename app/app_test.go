package app

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/Canto-Network/Canto-Testnet-v2/v1/types"
	"github.com/Canto-Network/ethermint-v2/encoding"
)

func TestcantoExport(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewCanto(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encoding.MakeConfig(ModuleBasics), simapp.EmptyAppOptions{})

	genesisState := NewDefaultGenesisState()
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			ChainId:       types.MainnetChainID + "-1",
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewCanto(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encoding.MakeConfig(ModuleBasics), simapp.EmptyAppOptions{})
	_, err = app2.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
