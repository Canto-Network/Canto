package testutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/Canto-Network/Canto/v6/testutil/network"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/client/cli"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

var fundAccount sdk.AccAddress

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.T().Log("setting up integration test suite")
	cfg := network.DefaultConfig()
	cfg.NumValidators = 1
	// Used "stake" as denom because bonded denom was set in DefaultConfig() by using
	// app.ModuleBasics.DefaultGenesis(encCfg.Marshaler).
	// TODO: We need to update that DefaultConfig() to use "acanto" as bonded denom.
	cfg.BondDenom = sdk.DefaultBondDenom
	cfg.MinGasPrices = fmt.Sprintf("0.0001%s", cfg.BondDenom)
	cfg.AccountTokens = types.ChunkSize.MulRaw(10000)
	cfg.StakingTokens = types.ChunkSize.MulRaw(5000)
	cfg.BondedTokens = types.ChunkSize.MulRaw(1000)
	suite.cfg = cfg

	// genStateLiquidStaking := types.DefaultGenesisState()
	path, err := os.MkdirTemp("/tmp", "lct-*")
	suite.NoError(err)
	suite.network, err = network.New(suite.T(), path, suite.cfg)
	suite.NoError(err)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.T().Log("tearing down integration test suite")
	suite.network.Cleanup()
}

func (suite *IntegrationTestSuite) TestCmdParams() {
	val := suite.network.Validators[0]

	tcs := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"dynamic_fee_rate":{"r0":"0.000000000000000000","u_soft_cap":"0.050000000000000000","u_hard_cap":"0.100000000000000000","u_optimal":"0.090000000000000000","slope1":"0.100000000000000000","slope2":"0.400000000000000000","max_fee_rate":"0.500000000000000000"}}`,
		},
		// TODO: output flag is set to text, but output is still json
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`{"dynamic_fee_rate":{"r0":"0.000000000000000000","u_soft_cap":"0.050000000000000000","u_hard_cap":"0.100000000000000000","u_optimal":"0.090000000000000000","slope1":"0.100000000000000000","slope2":"0.400000000000000000","max_fee_rate":"0.500000000000000000"}}`,
		},
	}
	for _, tc := range tcs {
		tc := tc

		suite.Run(tc.name, func() {
			cmd := cli.CmdQueryParams()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			suite.Require().NoError(err)
			suite.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
		})
	}
}

// TestLiquidStaking includes testing for
// * CmdQueryStates
func (suite *IntegrationTestSuite) TestLiquidStaking() {
	vals := suite.network.Validators
	clientCtx := vals[0].ClientCtx
	states := suite.getStates(clientCtx)
	suite.True(states.IsZeroState())

	minCollateral := suite.getMinimumCollateral(clientCtx)
	// provide 3 insurances
	tenPercent := sdk.NewDecWithPrec(10, 2)
	oneInsuranceAmt := minCollateral.Amount.TruncateInt()
	for i := 0; i < 3; i++ {
		_, err := ExecMsgProvideInsurance(
			clientCtx,
			vals[0].Address.String(),
			vals[0].ValAddress.String(),
			sdk.NewCoin(suite.cfg.BondDenom, oneInsuranceAmt),
			tenPercent,
		)
		suite.NoError(err)
	}
	suite.NoError(suite.network.WaitForNextBlock())
	states = suite.getStates(clientCtx)
	suite.True(
		states.TotalInsuranceTokens.Equal(oneInsuranceAmt.MulRaw(3)),
		"3 insurances are provided so total insurance tokens should be 3",
	)
	// Cancel 1 insurance, (3 -> 2)
	_, err := ExecMsgCancelProvideInsurance(
		clientCtx,
		vals[0].Address.String(),
		3,
	)
	suite.NoError(err)
	suite.NoError(suite.network.WaitForNextBlock())
	states = suite.getStates(clientCtx)
	suite.True(
		states.TotalInsuranceTokens.Equal(oneInsuranceAmt.MulRaw(2)),
		"1 insurance is canceled so total insurance tokens should be 2",
	)

	// liquid stake 2 chunks
	_, err = ExecMsgLiquidStake(
		clientCtx,
		vals[0].Address.String(),
		sdk.NewCoin(suite.cfg.BondDenom, types.ChunkSize.MulRaw(2)),
	)
	suite.NoError(err)
	suite.NoError(suite.network.WaitForNextBlock())
	states = suite.getStates(clientCtx)
	fmt.Println(states.RemainingChunkSlots.String())
	suite.True(states.Equal(types.NetAmountState{
		MintRate:                           sdk.OneDec(),
		LsTokensTotalSupply:                types.ChunkSize.MulRaw(2),
		NetAmount:                          types.ChunkSize.MulRaw(2).ToDec(),
		TotalLiquidTokens:                  types.ChunkSize.MulRaw(2),
		RewardModuleAccBalance:             sdk.ZeroInt(),
		FeeRate:                            sdk.ZeroDec(),
		UtilizationRatio:                   sdk.MustNewDecFromStr("0.0004"),
		RemainingChunkSlots:                sdk.NewInt(498),
		NumPairedChunks:                    sdk.NewInt(2),
		DiscountRate:                       sdk.ZeroDec(),
		TotalDelShares:                     types.ChunkSize.MulRaw(2).ToDec(),
		TotalRemainingRewards:              sdk.ZeroDec(),
		TotalChunksBalance:                 sdk.ZeroInt(),
		TotalUnbondingChunksBalance:        sdk.ZeroInt(),
		TotalInsuranceTokens:               oneInsuranceAmt.MulRaw(2),
		TotalPairedInsuranceTokens:         oneInsuranceAmt.MulRaw(2),
		TotalUnpairingInsuranceTokens:      sdk.ZeroInt(),
		TotalRemainingInsuranceCommissions: sdk.ZeroDec(),
	}))

	// liquid unstake 1 chunk
	_, err = ExecMsgLiquidUnstake(
		clientCtx,
		vals[0].Address.String(),
		sdk.NewCoin(suite.cfg.BondDenom, types.ChunkSize),
	)
	suite.NoError(err)
	suite.NoError(suite.network.WaitForNextBlock())

	infos := suite.getUnpairingForUnstakingChunkInfos(clientCtx, vals[0].Address.String())
	suite.Require().Len(infos, 1)
	suite.Equal(vals[0].Address.String(), infos[0].DelegatorAddress)
	suite.Equal(sdk.NewCoin(types.DefaultLiquidBondDenom, types.ChunkSize), infos[0].EscrowedLstokens)

	// TODO: how to implement advance blocks with time?
}

func (suite *IntegrationTestSuite) getMinimumCollateral(ctx client.Context) sdk.DecCoin {
	var res types.QueryMinimumCollateralResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryMinimumCollateral(),
		[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.MinimumCollateral
}

func (suite *IntegrationTestSuite) getUnpairingForUnstakingChunkInfos(
	ctx client.Context,
	delegator string,
) []types.UnpairingForUnstakingChunkInfo {
	var res types.QueryUnpairingForUnstakingChunkInfosResponse
	extraArgs := []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	if delegator != "" {
		extraArgs = append(extraArgs, fmt.Sprintf("--%s=%s", cli.FlagDelegatorAddress, delegator))
	}
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryUnpairingForUnstakingChunkInfos(),
		extraArgs,
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.UnpairingForUnstakingChunkInfos
}

func (suite *IntegrationTestSuite) getStates(ctx client.Context) types.NetAmountState {
	var res types.QueryStatesResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryStates(),
		[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.NetAmountState
}

// add test addresses with funds
func (suite *IntegrationTestSuite) addTestAddrsWithFunding(ctx client.Context, fundingAccount sdk.AccAddress, accNum int, amount sdk.Coin) ([]sdk.AccAddress, []sdk.Coin) {
	addrs := make([]sdk.AccAddress, 0, accNum)
	balances := make([]sdk.Coin, 0, accNum)
	for i := 0; i < accNum; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		addrs = append(addrs, addr)
		balances = append(balances, amount)

		_, err := ExecMsgSendCoins(ctx, fundingAccount.String(), addr.String(), sdk.NewCoins(amount))
		suite.NoError(err)
	}
	return addrs, balances
}

func (suite *IntegrationTestSuite) fundAccount(ctx client.Context, fundAccount, addr string, amount sdk.Coins) error {
	_, err := ExecMsgSendCoins(ctx, fundAccount, addr, amount)
	return err
}
