package testutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Canto-Network/Canto/v6/testutil/network"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/client/cli"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/client"
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

// SetupSuite setup some basic states to tests queries
func (suite *IntegrationTestSuite) SetupSuite() {
	suite.T().Log("setting up integration test suite")
	cfg := network.DefaultConfig()
	cfg.NumValidators = 1
	// Used "stake" as denom because bonded denom was set in DefaultConfig() by using
	// app.ModuleBasics.DefaultGenesis(encCfg.Marshaler).
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

func (suite *IntegrationTestSuite) TestCmdQueryParams() {
	val := suite.network.Validators[0]

	tcs := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"params":{"dynamic_fee_rate":{"r0":"0.000000000000000000","u_soft_cap":"0.050000000000000000","u_hard_cap":"0.100000000000000000","u_optimal":"0.090000000000000000","slope1":"0.100000000000000000","slope2":"0.400000000000000000","max_fee_rate":"0.500000000000000000"}}}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`params:
  dynamic_fee_rate:
    max_fee_rate: "0.500000000000000000"
    r0: "0.000000000000000000"
    slope1: "0.100000000000000000"
    slope2: "0.400000000000000000"
    u_hard_cap: "0.100000000000000000"
    u_optimal: "0.090000000000000000"
    u_soft_cap: "0.050000000000000000"
`,
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

func (suite *IntegrationTestSuite) TestCmdQueryChunkSize() {
	val := suite.network.Validators[0]

	tcs := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"chunk_size":{"denom":"stake","amount":"250000000000000000000000"}}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`chunk_size:
  amount: "250000000000000000000000"
  denom: stake
`,
		},
	}
	for _, tc := range tcs {
		tc := tc

		suite.Run(tc.name, func() {
			cmd := cli.CmdQueryChunkSize()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			suite.Require().NoError(err)
			suite.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
		})
	}
}

func (suite *IntegrationTestSuite) TestCmdQueryMinimumCollateral() {
	val := suite.network.Validators[0]

	tcs := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"minimum_collateral":{"denom":"stake","amount":"17500000000000000000000.000000000000000000"}}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`minimum_collateral:
  amount: "17500000000000000000000.000000000000000000"
  denom: stake
`,
		},
	}
	for _, tc := range tcs {
		tc := tc

		suite.Run(tc.name, func() {
			cmd := cli.CmdQueryMinimumCollateral()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			suite.Require().NoError(err)
			suite.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
		})
	}
}

// TestLiquidStaking tests liquidstaking module's actions by executing cmds.
// tested Txs:
// * CmdLiquidStake
// * CmdLiquidUnstake
// * CmdProvideInsurance
// * CmdCancelProvideInsurance
// * CmdDepositInsurance
// * CmdWithdrawInsurance
// * CmdWithdrawInsuranceCommission
// tested Queries:
// * CmdQueryEpoch
// * CmdQueryChunk
// * CmdQueryChunks
// * CmdQueryInsurance
// * CmdQueryInsurances
// * CmdQueryWithdrawInsuranceRequest
// * CmdQueryWithdrawInsuranceRequests
// * CmdQueryUnpairingForUnstakingChunkInfo
// * CmdQueryUnpairingForUnstakingChunkInfos
// * CmdQueryChunkSize
// * CmdQueryMinimumCollateral
// * CmdQueryStates
func (suite *IntegrationTestSuite) TestLiquidStaking() {
	vals := suite.network.Validators
	clientCtx := vals[0].ClientCtx

	epoch := suite.getEpoch(clientCtx)
	suite.Equal(stakingtypes.DefaultUnbondingTime, epoch.Duration)

	states := suite.getStates(clientCtx)
	suite.True(states.IsZeroState())

	minCollateral := suite.getMinimumCollateral(clientCtx)
	// provide an insurance
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
	insurances := suite.getAllInsurances(clientCtx)
	suite.Equal(3, len(insurances))
	for _, insurance := range insurances {
		result := suite.getInsurance(clientCtx, insurance.Id)
		suite.True(result.Equal(insurance))
	}
	states = suite.getStates(clientCtx)
	suite.True(
		states.TotalInsuranceTokens.Equal(oneInsuranceAmt.MulRaw(3)),
		"3 insurances are provided so total insurance tokens should be 3",
	)
	// Cancel 1 insurance
	_, err := ExecMsgCancelProvideInsurance(
		clientCtx,
		vals[0].Address.String(),
		3,
	)
	suite.NoError(err)
	states = suite.getStates(clientCtx)
	suite.True(
		states.TotalInsuranceTokens.Equal(oneInsuranceAmt.MulRaw(2)),
		"1 insurance is canceled so total insurance tokens should be 2",
	)

	// liquid stake 2 chunks
	for i := 0; i < 3; i++ {
		_, err = ExecMsgLiquidStake(
			clientCtx,
			vals[0].Address.String(),
			sdk.NewCoin(suite.cfg.BondDenom, types.ChunkSize),
		)
		suite.NoError(err)
	}
	chunks := suite.getAllChunks(clientCtx)
	suite.Equal(2, len(chunks))
	for _, chunk := range chunks {
		result := suite.getChunk(clientCtx, chunk.Id)
		suite.True(chunk.Equal(result))
	}
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

	infos := suite.getUnpairingForUnstakingChunkInfos(clientCtx, vals[0].Address.String())
	suite.Require().Len(infos, 1)
	suite.Equal(vals[0].Address.String(), infos[0].DelegatorAddress)
	for _, info := range infos {
		result := suite.getUnpairingForUnstakingChunkInfo(clientCtx, info.ChunkId)
		suite.True(info.Equal(result))
	}
	suite.Equal(sdk.NewCoin(types.DefaultLiquidBondDenom, types.ChunkSize), infos[0].EscrowedLstokens)

	// withdraw insurance commission
	_, err = ExecMsgWithdrawInsuranceCommission(clientCtx, vals[0].Address.String(), 2)
	suite.NoError(err)

	// Deposit insurance
	beforeBals := suite.getBalances(clientCtx, insurances[1].DerivedAddress())
	deposit := sdk.NewCoin(suite.cfg.BondDenom, sdk.NewInt(100))
	_, err = ExecMsgDepositInsurance(clientCtx, vals[0].Address.String(), 2, deposit)
	suite.NoError(err)
	afterBals := suite.getBalances(clientCtx, insurances[1].DerivedAddress())
	suite.Equal(
		afterBals.AmountOf(suite.cfg.BondDenom).Sub(beforeBals.AmountOf(suite.cfg.BondDenom)),
		deposit.Amount,
		"insurance should be deposited",
	)

	// withdraw insurance
	_, err = ExecMsgWithdrawInsurance(clientCtx, vals[0].Address.String(), 2)
	suite.NoError(err)

	reqs := suite.getWithdrawInsuranceRequests(clientCtx)
	suite.Require().Len(reqs, 1)
	for _, req := range reqs {
		result := suite.getWithdrawInsuranceRequest(clientCtx, req.InsuranceId)
		suite.True(req.Equal(result))
	}
}

func (suite *IntegrationTestSuite) getParams(ctx client.Context) types.Params {
	var res types.QueryParamsResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryParams(),
		[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.Params
}

// getStates returns all states by using cmdQueryStates
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

// getAllChunks returns all chunks by using cmdQueryChunks
func (suite *IntegrationTestSuite) getAllChunks(ctx client.Context) []types.Chunk {
	var res types.QueryChunksResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryChunks(),
		[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	var chunks []types.Chunk
	for _, chunk := range res.Chunks {
		chunks = append(chunks, chunk.Chunk)
	}
	return chunks
}

// getChunk returns a chunk with the given chunkID by using cmdQueryChunk
func (suite *IntegrationTestSuite) getChunk(ctx client.Context, chunkID uint64) types.Chunk {
	var res types.QueryChunkResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryChunk(),
		[]string{strconv.FormatUint(chunkID, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.Chunk
}

// getAllInsurances returns all insurances by using cmdQueryInsurances
func (suite *IntegrationTestSuite) getAllInsurances(ctx client.Context) []types.Insurance {
	var res types.QueryInsurancesResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryInsurances(),
		[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	var insurances []types.Insurance
	for _, ins := range res.Insurances {
		insurances = append(insurances, ins.Insurance)
	}
	return insurances
}

// getInsurance returns an insurance by using cmdQueryInsurance
func (suite *IntegrationTestSuite) getInsurance(ctx client.Context, insuranceId uint64) types.Insurance {
	var res types.QueryInsuranceResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryInsurance(),
		[]string{strconv.FormatUint(insuranceId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.Insurance
}

// getWithdrawInsuranceRequests returns all withdraw insurance requests by using cmdQueryWithdrawInsuranceRequests
func (suite *IntegrationTestSuite) getWithdrawInsuranceRequests(ctx client.Context) []types.WithdrawInsuranceRequest {
	var res types.QueryWithdrawInsuranceRequestsResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryWithdrawInsuranceRequests(),
		[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.WithdrawInsuranceRequests
}

// getWithdrawInsuranceRequest returns withdraw insurance request by using cmdQueryWithdrawInsuranceRequest
func (suite *IntegrationTestSuite) getWithdrawInsuranceRequest(ctx client.Context, insuranceId uint64) types.WithdrawInsuranceRequest {
	var res types.QueryWithdrawInsuranceRequestResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryWithdrawInsuranceRequest(),
		[]string{strconv.FormatUint(insuranceId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.WithdrawInsuranceRequest
}

// getUnpairingForUnstakingChunkInfos returns all unpairing for unstaking chunk infos by using cmdQueryUnpairingForUnstakingChunkInfos
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

// getUnpairingForUnstakingChunkInfo returns the unpairing for unstaking chunk info by using cmdQueryUnpairingForUnstakingChunkInfo
func (suite *IntegrationTestSuite) getUnpairingForUnstakingChunkInfo(ctx client.Context, chunkId uint64) types.UnpairingForUnstakingChunkInfo {
	var res types.QueryUnpairingForUnstakingChunkInfoResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryUnpairingForUnstakingChunkInfo(),
		[]string{strconv.FormatUint(chunkId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.UnpairingForUnstakingChunkInfo
}

// getMinimumCollateral returns minimum collateral by using cmdQueryMinimumCollateral
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

// getEpoch returns epoch by using cmdQueryEpoch
func (suite *IntegrationTestSuite) getEpoch(ctx client.Context) types.Epoch {
	var res types.QueryEpochResponse
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		cli.CmdQueryEpoch(),
		[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
	)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.Epoch
}

func (suite *IntegrationTestSuite) getBalances(ctx client.Context, addr sdk.AccAddress) sdk.Coins {
	var res banktypes.QueryAllBalancesResponse
	args := []string{addr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	out, err := clitestutil.ExecTestCLICmd(ctx, bankcli.GetBalancesCmd(), args)
	suite.NoError(err)
	suite.NoError(suite.cfg.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
	return res.Balances
}
