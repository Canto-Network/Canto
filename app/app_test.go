package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-db"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/state"
	cbfttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	"cosmossdk.io/store/iavl"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/Canto-Network/Canto/v8/types"
	coinswaptypes "github.com/Canto-Network/Canto/v8/x/coinswap/types"
	epochstypes "github.com/Canto-Network/Canto/v8/x/epochs/types"
	inflationtypes "github.com/Canto-Network/Canto/v8/x/inflation/types"
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

// TestWorkingHash tests that the working hash of the IAVL store is calculated correctly during the initialization phase of the genesis, given the initial height specified in the GenesisDoc.
func TestWorkingHash(t *testing.T) {
	gdoc, err := state.MakeGenesisDocFromFile("height4-genesis.json")
	require.NoError(t, err)

	gs, err := state.MakeGenesisState(gdoc)
	require.NoError(t, err)

	tmpDir := fmt.Sprintf("test-%s", time.Now().String())
	db, err := dbm.NewGoLevelDB("test", tmpDir, nil)
	require.NoError(t, err)
	app := NewCanto(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, false, simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome), baseapp.SetChainID("canto_9000-1"))

	// delete tmpDir
	defer require.NoError(t, os.RemoveAll(tmpDir))

	pbparams := gdoc.ConsensusParams.ToProto()
	// Initialize the chain
	_, err = app.InitChain(&abci.RequestInitChain{
		Time:            gdoc.GenesisTime,
		ChainId:         gdoc.ChainID,
		ConsensusParams: &pbparams,
		Validators:      cbfttypes.TM2PB.ValidatorUpdates(gs.Validators),
		AppStateBytes:   gdoc.AppState,
		InitialHeight:   gdoc.InitialHeight,
	})
	require.NoError(t, err)

	// Call FinalizeBlock to calculate each module's working hash.
	// Without calling this, all module's root node will have empty hash.
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: gdoc.InitialHeight,
		Time:   time.Now(),
	})
	require.NoError(t, err)

	storeKeys := app.GetStoreKeys()
	// deterministicKeys are module keys which has always same working hash whenever run this test. (non deterministic module: staking, epoch, inflation)
	deterministicKeys := []string{
		authtypes.StoreKey, banktypes.StoreKey, capabilitytypes.StoreKey, coinswaptypes.StoreKey,
		consensustypes.StoreKey, crisistypes.StoreKey, distrtypes.StoreKey,
		evmtypes.StoreKey, feemarkettypes.StoreKey, govtypes.StoreKey, ibctransfertypes.StoreKey,
		paramstypes.StoreKey, slashingtypes.StoreKey, upgradetypes.StoreKey}
	// workingHashWithZeroInitialHeight is the working hash of the IAVL store with initial height 0 with given genesis.
	// you can get this hash by running the test with iavl v1.1.2 with printing working hash (ref. https://github.com/b-harvest/cosmos-sdk/commit/4f44d6a2fe80ee7fe8c4409b820226e3615c6500)
	workingHashWithZeroInitialHeight := map[string]string{
		authtypes.StoreKey:        "939df0607b42b9dc73166475d95b34acc46eb7700c773dfb4936877ed545f52a",
		banktypes.StoreKey:        "149427e950d1ab4ee88a2028b52930f2b15ce89483b4640ce2dc93f11965045f",
		capabilitytypes.StoreKey:  "379646e1b21b0d39607965b0ad2371f83eece289eb469d024dd8ffc3cfbb0cd0",
		coinswaptypes.StoreKey:    "33e007488bc655cb90eb3a0ade1ac0168eb0d9bb3dd369f0cd05783f1afc8689",
		consensustypes.StoreKey:   "35760e4a68fbc1cdd5b3b181b90d04f51390a0aa55476cdb40924d1494bf3d1d",
		crisistypes.StoreKey:      "0244e5ce8733dda116da9b110abe8e4624bed97cb0d9342d685bfff3b7bec819",
		distrtypes.StoreKey:       "6da642a5ad9983f3d1590e950721d69e59611faadc2a5aaa46d462979658b743",
		epochstypes.StoreKey:      "afb6c6fdd19b56748393f6a0d07e07179b166c25e6f44fcd688c6d1dcb53b022",
		evmtypes.StoreKey:         "e5f6f682b247cd9812e05db063d355c3f2661aa12d03cf0b900879a1980d002b",
		feemarkettypes.StoreKey:   "986b2ae7d92fa552688e4370ae58009e7c2150d41741aae47e748de18bb49f67",
		govtypes.StoreKey:         "8e160b59f63d310da86e394c68deb47aa2cc47d61afec94b8677d6fb660dae43",
		ibctransfertypes.StoreKey: "90fdae9ef2124b88045deea846624eed6bc9e5e5422fada847b882fb66d81f17",
		inflationtypes.StoreKey:   "2f694d43447679f2e058e5629d3305727566e5870dbea6010c574efa8e67ea8e",
		paramstypes.StoreKey:      "da7007ea1f6f90372c4bf724f30e024a95980353c09b5ed5272de30411604e53",
		slashingtypes.StoreKey:    "3abe3df91a605993b74446f58c82f17ae1dbebd155cb474c01972befa1e1ca03",
		stakingtypes.StoreKey:     "48c4ec05af62a8171124c88c2aeca67ca281a71df7047767ba0004d82e80ad84",
		upgradetypes.StoreKey:     "530a41d3858a8f8fb48ec710e543d633ff0b9ab71500f7c98a13515e53a0f6cd",
	}
	// workingHashWithCorrectInitialHeight is the working hash of the IAVL store with correct initial height 4 with given genesis.
	workingHashWithCorrectInitialHeight := map[string]string{
		authtypes.StoreKey:        "b513a8bc783341508209bef42f2ca0d97f2e12e3f977e494d30ccd06ba2049f3",
		banktypes.StoreKey:        "cf0406a0e743fd4297d67b80dc245c81c466d722c8f2415c97b892d1da592c19",
		capabilitytypes.StoreKey:  "e9261548b1c687638721f75920e1c8e3f4f52cbd7ab3aeddc6c626cd8abc8718",
		coinswaptypes.StoreKey:    "57094d1ec4776eab36ae55145557a8679dca01529feed5e83bb1753a8849ed28",
		consensustypes.StoreKey:   "2573460b2ef3a08c825bdfb485e51680038530f70c45f0b723167ae09599761c",
		crisistypes.StoreKey:      "7c169f2c8fa4c6f64a61154e4fb3ffb3d74893a8bf6efdf50fc6e06d92979ecb",
		distrtypes.StoreKey:       "65cfb5c307b023ed80255b3ffc14e08b39cc00ecd7b710ff8b5bc96d1efdbda6",
		epochstypes.StoreKey:      "a26294cd8405c0e3cabc508c90d34317bfe547c90fd4be22223d27aecd819211",
		evmtypes.StoreKey:         "f2dcba601044394f83455be931d2ad78a08ede45873b16d8884d5f650db42f99",
		feemarkettypes.StoreKey:   "b8a66cce8e7809f521db9fdd71bfeb980966f1b74ef252bda65804d8b89da7de",
		govtypes.StoreKey:         "033e4a113195025eb54ecdeabffec5fd20605eae08da125d237768f1f3387616",
		ibctransfertypes.StoreKey: "3ffd548eb86288efc51964649e36dc710f591c3d60d6f9c1b42f2a4d17870904",
		inflationtypes.StoreKey:   "b85bc597af9eb62c42e06b0f158bde591975585bab384e9666f89f80001b3d01",
		paramstypes.StoreKey:      "689664bc7d4f9aa8abeeb6ba13a556467e149d56e7bead89e8b574bdf8e0fe7f",
		slashingtypes.StoreKey:    "9da3ff2ded57e30dfea0371278d9043bea9f579421beb45b58ec7240e1b4f27a",
		stakingtypes.StoreKey:     "17b36186121d21b713a667c6cd534562bbc3095974ec866cb906cad7c2fa31e1",
		upgradetypes.StoreKey:     "9677219870ca98ba9868589ccdcd97411d9b82abc6a7aa1910016457b4730a93",
	}

	matchAny := func(key string) bool {
		for _, dk := range deterministicKeys {
			if dk == key {
				return true
			}
		}
		return false
	}

	for _, key := range storeKeys {
		if key != nil && matchAny(key.Name()) {
			kvstore := app.CommitMultiStore().GetCommitKVStore(key)
			require.Equal(t, storetypes.StoreTypeIAVL, kvstore.GetStoreType())
			iavlStore, ok := kvstore.(*iavl.Store)
			require.True(t, ok)
			workingHash := hex.EncodeToString(iavlStore.WorkingHash())
			require.NotEqual(t, workingHashWithZeroInitialHeight[key.Name()], workingHash)
			require.Equal(t, workingHashWithCorrectInitialHeight[key.Name()], workingHash, key.Name())
		}
	}
}
