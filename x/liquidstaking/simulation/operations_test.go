package simulation_test

import (
	"github.com/Canto-Network/Canto/v7/app/params"
	"math/rand"
	"testing"
	"time"

	"github.com/Canto-Network/Canto/v7/app"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/simulation"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestWeightedOperations(t *testing.T) {
	canto, ctx := createTestApp(false)
	cdc := types.ModuleCdc
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(
		appParams,
		cdc,
		canto.AccountKeeper,
		canto.BankKeeper,
		canto.StakingKeeper,
		canto.LiquidStakingKeeper,
	)

	s := rand.NewSource(2)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, canto, ctx, 10)

	// setup accounts[0] as validator0 and accounts[1] as validator1
	getTestingValidator0(t, canto, ctx, accs)
	getTestingValidator1(t, canto, ctx, accs)

	blockTime := time.Now().UTC()
	canto.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  canto.LastBlockHeight() + 1,
			AppHash: canto.LastCommitID().Hash,
			Time:    blockTime,
		},
	})
	canto.EndBlock(abci.RequestEndBlock{Height: canto.LastBlockHeight() + 1})

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{params.DefaultWeightMsgLiquidStake, types.ModuleName, types.TypeMsgLiquidStake},
		{params.DefaultWeightMsgLiquidUnstake, types.ModuleName, types.TypeMsgLiquidUnstake},
		{params.DefaultWeightMsgProvideInsurance, types.ModuleName, types.TypeMsgProvideInsurance},
		{params.DefaultWeightMsgCancelProvideInsurance, types.ModuleName, types.TypeMsgCancelProvideInsurance},
		{params.DefaultWeightMsgDepositInsurance, types.ModuleName, types.TypeMsgDepositInsurance},
		{params.DefaultWeightMsgWithdrawInsurance, types.ModuleName, types.TypeMsgWithdrawInsurance},
		{params.DefaultWeightMsgWithdrawInsuranceCommission, types.ModuleName, types.TypeMsgWithdrawInsuranceCommission},
		{params.DefaultWeightMsgClaimDiscountedReward, types.ModuleName, types.TypeMsgClaimDiscountedReward},
	}

	for i, w := range weightedOps {
		opMsg, _, _ := w.Op()(r, canto.BaseApp, ctx, accs, ctx.ChainID())
		require.Equal(t, expected[i].weight, w.Weight())
		require.Equal(t, expected[i].opMsgRoute, opMsg.Route)
		require.Equal(t, expected[i].opMsgName, opMsg.Name)
	}
}

func createTestApp(isCheckTx bool) (*app.Canto, sdk.Context) {
	app := app.Setup(isCheckTx, nil)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
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

func fundAccount(bk types.BankKeeper, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	if err := bk.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}
	if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins); err != nil {
		return err
	}
	return nil
}

func getTestingValidator0(t *testing.T, app *app.Canto, ctx sdk.Context, accounts []simtypes.Account) stakingtypes.Validator {
	commission0 := stakingtypes.NewCommission(sdk.ZeroDec(), sdk.OneDec(), sdk.OneDec())
	return getTestingValidator(t, app, ctx, accounts, commission0, 0)
}

func getTestingValidator1(t *testing.T, app *app.Canto, ctx sdk.Context, accounts []simtypes.Account) stakingtypes.Validator {
	commission1 := stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	return getTestingValidator(t, app, ctx, accounts, commission1, 1)
}

func getTestingValidator(t *testing.T, app *app.Canto, ctx sdk.Context, accounts []simtypes.Account, commission stakingtypes.Commission, n int) stakingtypes.Validator {
	account := accounts[n]
	valPubKey := account.PubKey
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator := teststaking.NewValidator(t, valAddr, valPubKey)
	validator, err := validator.SetInitialCommission(commission)
	require.NoError(t, err)

	validator.DelegatorShares = sdk.NewDec(100)
	validator.Tokens = app.StakingKeeper.TokensFromConsensusPower(ctx, 100)

	app.StakingKeeper.SetValidator(ctx, validator)

	return validator
}
