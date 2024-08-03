package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Canto-Network/Canto/v8/app/params"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v8/app"
	"github.com/Canto-Network/Canto/v8/x/govshuttle/simulation"
	"github.com/Canto-Network/Canto/v8/x/govshuttle/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

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
	ctx = ctx.WithChainID("canto_9001-1")
	return app, ctx
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

func TestProposalMsgs(t *testing.T) {
	app, ctx := createTestApp(t, false)

	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs(app.GovshuttleKeeper)
	require.Equal(t, 2, len(weightedProposalMsgs))

	w0 := weightedProposalMsgs[0]
	w1 := weightedProposalMsgs[1]

	// tests w0 interface
	require.Equal(t, simulation.OpWeightSimulateLendingMarketProposal, w0.AppParamsKey())
	require.Equal(t, params.DefaultWeightLendingMarketProposal, w0.DefaultWeight())

	// tests w1 interface
	require.Equal(t, simulation.OpWeightSimulateTreasuryProposal, w1.AppParamsKey())
	require.Equal(t, params.DefaultWeightTreasuryProposal, w1.DefaultWeight())

	msg := w0.MsgSimulatorFn()(r, ctx, accounts)
	MsgLendingMarket, ok := msg.(*types.MsgLendingMarketProposal)
	require.True(t, ok)
	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), MsgLendingMarket.Authority)

	msg = w1.MsgSimulatorFn()(r, ctx, accounts)
	MsgTreasury, ok := msg.(*types.MsgTreasuryProposal)
	require.True(t, ok)
	require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), MsgTreasury.Authority)

}
