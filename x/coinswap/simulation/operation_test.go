package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v8/app/params"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/Canto-Network/Canto/v8/app"
	"github.com/Canto-Network/Canto/v8/x/coinswap/simulation"
	"github.com/Canto-Network/Canto/v8/x/coinswap/types"
)

func TestWeightedOperations(t *testing.T) {
	canto, ctx := createTestApp(t, false)
	cdc := types.ModuleCdc
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(
		appParams,
		cdc,
		canto.CoinswapKeeper,
		canto.AccountKeeper,
		canto.BankKeeper,
	)

	s := rand.NewSource(2)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, canto, ctx, 10)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{params.DefaultWeightMsgAddLiquidity, types.ModuleName, sdk.MsgTypeURL(&types.MsgAddLiquidity{})},
		{params.DefaultWeightMsgSwapOrder, types.ModuleName, sdk.MsgTypeURL(&types.MsgSwapOrder{})},
		{params.DefaultWeightMsgRemoveLiquidity, types.ModuleName, sdk.MsgTypeURL(&types.MsgRemoveLiquidity{})},
	}

	for i, w := range weightedOps {
		opMsg, _, _ := w.Op()(r, canto.BaseApp, ctx, accs, ctx.ChainID())
		require.Equal(t, expected[i].weight, w.Weight())
		require.Equal(t, expected[i].opMsgRoute, opMsg.Route)
		require.Equal(t, expected[i].opMsgName, opMsg.Name)
	}
}

func createTestApp(t *testing.T, isCheckTx bool) (*app.Canto, sdk.Context) {
	app := app.Setup(isCheckTx, nil)
	r := rand.New(rand.NewSource(1))

	simAccs := simtypes.RandomAccounts(r, 10)

	ctx := app.BaseApp.NewContext(isCheckTx)
	validator := getTestingValidator0(t, app, ctx, simAccs)
	consAddr, err := validator.GetConsAddr()
	require.NoError(t, err)
	ctx = ctx.WithBlockHeader(cmtproto.Header{Height: 1,
		ChainID:         "canto_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddr,
	})
	return app, ctx
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *app.Canto, ctx sdk.Context, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 100_000_000)
	initCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, initAmt),
	)

	// add coins to the accounts
	for _, account := range accounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		err := fundAccount(app.BankKeeper, ctx, account.Address, initCoins)
		require.NoError(t, err)
	}

	return accounts
}

func fundAccount(bk bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	if err := bk.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}
	if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins); err != nil {
		return err
	}
	return nil
}

func getTestingValidator0(t *testing.T, app *app.Canto, ctx sdk.Context, accounts []simtypes.Account) stakingtypes.Validator {
	commission0 := stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec())
	return getTestingValidator(t, app, ctx, accounts, commission0, 0)
}

func getTestingValidator(t *testing.T, app *app.Canto, ctx sdk.Context, accounts []simtypes.Account, commission stakingtypes.Commission, n int) stakingtypes.Validator {
	account := accounts[n]
	valPubKey := account.PubKey
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator := testutil.NewValidator(t, valAddr, valPubKey)
	validator, err := validator.SetInitialCommission(commission)
	require.NoError(t, err)

	validator.DelegatorShares = sdkmath.LegacyNewDec(100)
	validator.Tokens = app.StakingKeeper.TokensFromConsensusPower(ctx, 100)

	app.StakingKeeper.SetValidator(ctx, validator)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)

	return validator
}
