package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/evmos/ethermint/encoding"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	epochstypes "github.com/Canto-Network/Canto/v6/x/epochs/types"
	liquidstakingtypes "github.com/Canto-Network/Canto/v6/x/liquidstaking/types"

	cantoconfig "github.com/Canto-Network/Canto/v6/cmd/config"
	inflationtypes "github.com/Canto-Network/Canto/v6/x/inflation/types"
)

// Get flags every time the simulator is run
func init() {
	simapp.GetSimulatorFlags()
}

type StoreKeysPrefixes struct {
	A        sdk.StoreKey
	B        sdk.StoreKey
	Prefixes [][]byte
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(app *baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestFullAppSimulation(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-cantoApp-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}
	config.ChainID = "canto_9000-1"
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	// TODO: shadowed
	cantoApp := NewCanto(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, simapp.FlagPeriodValue,
		encoding.MakeConfig(ModuleBasics), EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, cantoconfig.AppName, cantoApp.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		cantoApp.BaseApp,
		AppStateFn(cantoApp.AppCodec(), cantoApp.SimulationManager()),
		RandomAccounts, // replace with own random account function if using keys other than secp256k1
		simapp.SimulationOperations(cantoApp, cantoApp.AppCodec(), config),
		cantoApp.ModuleAccountAddrs(),
		config,
		cantoApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(cantoApp, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}
}

func TestAppImportExport(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	config.ChainID = "canto_9000-1"
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewCanto(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		simapp.FlagPeriodValue,
		encoding.MakeConfig(ModuleBasics),
		EmptyAppOptions{},
		fauxMerkleModeOpt,
	)
	require.Equal(t, cantoconfig.AppName, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		RandomAccounts, // replace with own random account function if using keys other than secp256k1
		simapp.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}

	fmt.Println("exporting genesis...")

	exported, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	fmt.Println("importing genesis...")

	_, newDB, newDir, _, _, err := simapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewCanto(
		log.NewNopLogger(),
		newDB,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		simapp.FlagPeriodValue,
		encoding.MakeConfig(ModuleBasics),
		EmptyAppOptions{},
		fauxMerkleModeOpt,
	)
	require.Equal(t, cantoconfig.AppName, newApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	newApp.mm.InitGenesis(ctxB, app.AppCodec(), genesisState)
	newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)

	fmt.Println("comparing stores...")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keys[authtypes.StoreKey], newApp.keys[authtypes.ModuleName], [][]byte{}},
		{app.keys[banktypes.StoreKey], newApp.keys[banktypes.ModuleName], [][]byte{banktypes.BalancesPrefix}},
		{app.keys[stakingtypes.StoreKey], newApp.keys[stakingtypes.ModuleName],
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey,
			},
		},
		{app.keys[distrtypes.StoreKey], newApp.keys[distrtypes.ModuleName], [][]byte{}},
		{app.keys[paramstypes.StoreKey], newApp.keys[paramstypes.ModuleName], [][]byte{}},
		{app.keys[upgradetypes.StoreKey], newApp.keys[upgradetypes.ModuleName], [][]byte{}},
		{app.keys[evidencetypes.StoreKey], newApp.keys[evidencetypes.ModuleName], [][]byte{}},
		{app.keys[capabilitytypes.StoreKey], newApp.keys[capabilitytypes.ModuleName], [][]byte{}},
		//{app.keys[feegrant.StoreKey], newApp.keys[feegrant.ModuleName], [][]byte{}},
		{app.keys[authzkeeper.StoreKey], newApp.keys[authz.ModuleName], [][]byte{}},
		{app.keys[ibchost.StoreKey], newApp.keys[ibchost.ModuleName], [][]byte{}},
		{app.keys[ibctransfertypes.StoreKey], newApp.keys[ibctransfertypes.ModuleName], [][]byte{}},
		//{app.keys[evmtypes.StoreKey], newApp.keys[evmtypes.ModuleName], [][]byte{}},
		{app.keys[feemarkettypes.StoreKey], newApp.keys[feemarkettypes.ModuleName], [][]byte{}},
		{app.keys[inflationtypes.StoreKey], newApp.keys[inflationtypes.ModuleName], [][]byte{}},
		//{app.keys[erc20types.StoreKey], newApp.keys[erc20types.ModuleName], [][]byte{}},
		{app.keys[epochstypes.StoreKey], newApp.keys[epochstypes.ModuleName], [][]byte{}},
		//{app.keys[vestingtypes.StoreKey], newApp.keys[vestingtypes.ModuleName], [][]byte{}},
		//{app.keys[recoverytypes.StoreKey], newApp.keys[recoverytypes.ModuleName], [][]byte{}},
		//{app.keys[feestypes.StoreKey], newApp.keys[feestypes.ModuleName], [][]byte{}},
		//{app.keys[csrtypes.StoreKey], newApp.keys[csrtypes.ModuleName], [][]byte{}},
		//{app.keys[govshuttletypes.StoreKey], newApp.keys[govshuttletypes.ModuleName], [][]byte{}},
		{app.keys[liquidstakingtypes.StoreKey], newApp.keys[liquidstakingtypes.ModuleName], [][]byte{}},
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, len(failedKVAs), 0, simapp.GetSimulationLog(skp.A.Name(), app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

func TestAppStateDeterminism(t *testing.T) {
	if !simapp.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simapp.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = "canto_9000-1"

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simapp.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			app := NewCanto(
				logger,
				db,
				nil,
				true,
				map[int64]bool{},
				DefaultNodeHome,
				simapp.FlagPeriodValue,
				encoding.MakeConfig(ModuleBasics),
				EmptyAppOptions{},
				fauxMerkleModeOpt,
			)
			fmt.Printf("running simulation with seed %d\n", config.Seed)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				AppStateFn(app.AppCodec(), app.SimulationManager()),
				RandomAccounts,
				simapp.SimulationOperations(app, app.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simapp.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n",
					config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed)
			}
		}
	}
}
