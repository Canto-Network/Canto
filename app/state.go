package app

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

var FlagGenesisTimeValue = int64(1640995200)

// AppStateFn returns the initial application state using a genesis or the simulation parameters.
// It panics if the user provides files for both of them.
// If a file is not given for the genesis or the sim params, it creates a randomized one.
func AppStateFn(cdc codec.JSONCodec, simManager *module.SimulationManager) simtypes.AppStateFn {
	return func(r *rand.Rand, accs []simtypes.Account, config simtypes.Config,
	) (appState json.RawMessage, simAccs []simtypes.Account, chainID string, genesisTimestamp time.Time) {
		if FlagGenesisTimeValue == 0 {
			genesisTimestamp = simtypes.RandTimestamp(r)
		} else {
			genesisTimestamp = time.Unix(FlagGenesisTimeValue, 0)
		}

		chainID = config.ChainID
		switch {
		case config.ParamsFile != "" && config.GenesisFile != "":
			panic("cannot provide both a genesis file and a params file")

		case config.GenesisFile != "":
			// override the default chain-id from simapp to set it later to the config
			genesisDoc, accounts := AppStateFromGenesisFileFn(r, cdc, config.GenesisFile)

			if FlagGenesisTimeValue == 0 {
				// use genesis timestamp if no custom timestamp is provided (i.e no random timestamp)
				genesisTimestamp = genesisDoc.GenesisTime
			}

			appState = genesisDoc.AppState
			chainID = genesisDoc.ChainID
			simAccs = accounts

		case config.ParamsFile != "":
			appParams := make(simtypes.AppParams)
			bz, err := os.ReadFile(config.ParamsFile)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(bz, &appParams)
			if err != nil {
				panic(err)
			}
			appState, simAccs = AppStateRandomizedFn(simManager, r, cdc, accs, genesisTimestamp, appParams)

		default:
			appParams := make(simtypes.AppParams)
			appState, simAccs = AppStateRandomizedFn(simManager, r, cdc, accs, genesisTimestamp, appParams)
		}
		rawState := make(map[string]json.RawMessage)
		err := json.Unmarshal(appState, &rawState)
		if err != nil {
			panic(err)
		}

		stakingStateBz, ok := rawState[stakingtypes.ModuleName]
		if !ok {
			panic(fmt.Sprintf("%s genesis state not found", stakingtypes.ModuleName))
		}
		stakingState := new(stakingtypes.GenesisState)
		err = cdc.UnmarshalJSON(stakingStateBz, stakingState)
		if err != nil {
			panic(err)
		}
		// compute not bonded balance
		notBondedTokens := sdk.ZeroInt()
		for _, val := range stakingState.Validators {
			if val.Status != stakingtypes.Unbonded {
				continue
			}
			notBondedTokens = notBondedTokens.Add(val.GetTokens())
		}
		notBondedCoin := sdk.NewCoin(stakingState.Params.BondDenom, notBondedTokens)
		bankStateBz, ok := rawState[banktypes.ModuleName]
		if !ok {
			panic(fmt.Sprintf("%s genesis state not found", banktypes.ModuleName))
		}
		bankState := new(banktypes.GenesisState)
		err = cdc.UnmarshalJSON(bankStateBz, bankState)
		if err != nil {
			panic(err)
		}

		stakingAddr := authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String()
		var found bool
		for _, balance := range bankState.Balances {
			if balance.Address == stakingAddr {
				found = true
				break
			}
		}
		if !found {
			bankState.Balances = append(bankState.Balances, banktypes.Balance{
				Address: stakingAddr,
				Coins:   sdk.NewCoins(notBondedCoin),
			})
		}

		// chagnge appState back
		rawState[stakingtypes.ModuleName], err = cdc.MarshalJSON(stakingState)
		rawState[banktypes.ModuleName], err = cdc.MarshalJSON(bankState)

		// replace appstate
		appState, err = json.Marshal(rawState)
		if err != nil {
			panic(err)
		}
		return appState, simAccs, chainID, genesisTimestamp
	}
}

// AppStateRandomizedFn creates calls each module's GenesisState generator function
// and creates the simulation params
func AppStateRandomizedFn(
	simManager *module.SimulationManager, r *rand.Rand, cdc codec.JSONCodec,
	accs []simtypes.Account, genesisTimestamp time.Time, appParams simtypes.AppParams,
) (json.RawMessage, []simtypes.Account) {
	numAccs := int64(len(accs))
	genesisState := NewDefaultGenesisState()

	var initalStake, numInitiallyBonded int64
	appParams.GetOrGenerate(
		cdc,
		simappparams.StakePerAccount,
		&initalStake,
		r,
		func(r *rand.Rand) { initalStake = r.Int63n(1e12) },
	)
	appParams.GetOrGenerate(
		cdc,
		simappparams.InitiallyBondedValidators,
		&numInitiallyBonded,
		r,
		func(r *rand.Rand) { numInitiallyBonded = int64(r.Intn(300)) },
	)

	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}

	fmt.Printf(
		`Selected randomly generated parameters for simulated genesis:
{
  StakePerAccount: %d,
  InitiallyBondedValidators: %d,
}
`, initalStake, numInitiallyBonded)

	simState := &module.SimulationState{
		AppParams:    appParams,
		Cdc:          cdc,
		Rand:         r,
		GenState:     genesisState,
		Accounts:     accs,
		InitialStake: initalStake,
		NumBonded:    numInitiallyBonded,
		GenTimestamp: genesisTimestamp,
	}

	simManager.GenerateGenesisStates(simState)

	appState, err := json.Marshal(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs
}

// AppStateFromGenesisFileFn util function to generate the genesis AppState
// from a genesis.json file.
func AppStateFromGenesisFileFn(r io.Reader, cdc codec.JSONCodec, genesisFile string) (tmtypes.GenesisDoc, []simtypes.Account) {
	bytes, err := os.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}

	var genesis tmtypes.GenesisDoc
	if err := tmjson.Unmarshal(bytes, &genesis); err != nil {
		panic(err)
	}

	var appState GenesisState
	err = json.Unmarshal(genesis.AppState, &appState)
	if err != nil {
		panic(err)
	}

	var authGenesis authtypes.GenesisState
	if appState[authtypes.ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[authtypes.ModuleName], &authGenesis)
	}

	newAccs := make([]simtypes.Account, len(authGenesis.Accounts))
	for i, acc := range authGenesis.Accounts {
		privkeySeed := make([]byte, 15)
		if _, err := r.Read(privkeySeed); err != nil {
			panic(err)
		}
		privKey, err := ethsecp256k1.GenerateKey()
		if err != nil {
			panic(err)
		}

		a, ok := acc.GetCachedValue().(authtypes.AccountI)
		if !ok {
			panic("account error")
		}

		simAcc := simtypes.Account{
			PrivKey: privKey,
			PubKey:  privKey.PubKey(),
			Address: a.GetAddress(),
			ConsKey: ed25519.GenPrivKeyFromSecret(privkeySeed),
		}
		newAccs[i] = simAcc
	}

	return genesis, newAccs
}
