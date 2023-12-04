package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/evmos/ethermint/encoding"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"

	cantoconfig "github.com/Canto-Network/Canto/v7/cmd/config"
	csrtypes "github.com/Canto-Network/Canto/v7/x/csr/types"
	erc20types "github.com/Canto-Network/Canto/v7/x/erc20/types"
	govshuttletypes "github.com/Canto-Network/Canto/v7/x/govshuttle/types"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
)

// Get flags every time the simulator is run
func init() {
	simapp.GetSimulatorFlags()
}

type StoreKeysPrefixes struct {
	A        storetypes.StoreKey
	B        storetypes.StoreKey
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
		true, encoding.MakeConfig(ModuleBasics), simapp.EmptyAppOptions{}, fauxMerkleModeOpt)
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

	sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000)

	app := NewCanto(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		simapp.FlagPeriodValue,
		true,
		encoding.MakeConfig(ModuleBasics),
		simapp.EmptyAppOptions{},
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
		true,
		encoding.MakeConfig(ModuleBasics),
		simapp.EmptyAppOptions{},
		fauxMerkleModeOpt,
	)
	require.Equal(t, cantoconfig.AppName, newApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := app.NewContextLegacy(true, tmproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContextLegacy(true, tmproto.Header{ChainID: config.ChainID, Height: app.LastBlockHeight()})
	newApp.mm.InitGenesis(ctxB, app.AppCodec(), genesisState)
	newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)

	fmt.Println("comparing stores...")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keys[authtypes.StoreKey], newApp.keys[authtypes.StoreKey], [][]byte{}},
		{app.keys[banktypes.StoreKey], newApp.keys[banktypes.StoreKey], [][]byte{banktypes.BalancesPrefix}},
		{app.keys[stakingtypes.StoreKey], newApp.keys[stakingtypes.StoreKey],
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey,
			},
		},
		{app.keys[distrtypes.StoreKey], newApp.keys[distrtypes.StoreKey], [][]byte{}},
		{app.keys[paramstypes.StoreKey], newApp.keys[paramstypes.StoreKey], [][]byte{}},
		{app.keys[evidencetypes.StoreKey], newApp.keys[evidencetypes.StoreKey], [][]byte{}},
		{app.keys[capabilitytypes.StoreKey], newApp.keys[capabilitytypes.StoreKey], [][]byte{}},
		{app.keys[feegrant.StoreKey], newApp.keys[feegrant.StoreKey], [][]byte{}},
		{app.keys[authzkeeper.StoreKey], newApp.keys[authzkeeper.StoreKey], [][]byte{}},
		{app.keys[ibchost.StoreKey], newApp.keys[ibchost.StoreKey], [][]byte{}},
		{app.keys[ibctransfertypes.StoreKey], newApp.keys[ibctransfertypes.StoreKey], [][]byte{}},
		{app.keys[evmtypes.StoreKey], newApp.keys[evmtypes.StoreKey], [][]byte{}},
		{app.keys[feemarkettypes.StoreKey], newApp.keys[feemarkettypes.StoreKey], [][]byte{}},
		{app.keys[inflationtypes.StoreKey], newApp.keys[inflationtypes.StoreKey], [][]byte{}},
		{app.keys[erc20types.StoreKey], newApp.keys[erc20types.StoreKey], [][]byte{}},
		// In the case of epoch module, the value is updated when importing genesis, so the store consistency is broken
		//{app.keys[epochstypes.StoreKey], newApp.keys[epochstypes.StoreKey], [][]byte{}},
		{app.keys[csrtypes.StoreKey], newApp.keys[csrtypes.StoreKey], [][]byte{}},
		{app.keys[govshuttletypes.StoreKey], newApp.keys[govshuttletypes.StoreKey], [][]byte{}},
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

	numSeeds := config.NumBlocks / 10
	numTimesToRunPerSeed := 2
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000)

	for i := 0; i < numSeeds; i++ {
		config.Seed = config.Seed + int64(i)
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
				true,
				encoding.MakeConfig(ModuleBasics),
				simapp.EmptyAppOptions{},
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

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation after import")
	}
	require.NoError(t, err, "simulation setup failed")
	config.ChainID = "canto_9000-1"

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000)

	app := NewCanto(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		simapp.FlagPeriodValue,
		true,
		encoding.MakeConfig(ModuleBasics),
		simapp.EmptyAppOptions{},
		fauxMerkleModeOpt,
	)
	require.Equal(t, cantoconfig.AppName, app.Name())

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		RandomAccounts, // Replace with own random account function if using keys other than secp256k1
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

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(true, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

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
		true,
		encoding.MakeConfig(ModuleBasics),
		simapp.EmptyAppOptions{},
		fauxMerkleModeOpt,
	)
	require.Equal(t, cantoconfig.AppName, newApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctx := newApp.NewContextLegacy(true, tmproto.Header{ChainID: config.ChainID, Height: app.LastBlockHeight()})
	newApp.mm.InitGenesis(ctx, app.AppCodec(), genesisState)
	newApp.StoreConsensusParams(ctx, exported.ConsensusParams)

	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		newApp.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simapp.SimulationOperations(newApp, newApp.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)
	require.NoError(t, err)
}
