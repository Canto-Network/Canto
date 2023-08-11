package testutil

import (
	"fmt"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)).String()),
	fmt.Sprintf("--%s=%s", flags.FlagGas, "10000000"),
}

func ExecMsgProvideInsurance(clientCtx client.Context, from, validatorAddress string, amount sdk.Coin, feeRate sdk.Dec, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		validatorAddress,
		amount.String(),
		feeRate.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewProvideInsuranceCmd(), args)
}

func ExecMsgCancelProvideInsurance(clientCtx client.Context, from string, insuranceId uint64, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		fmt.Sprintf("%d", insuranceId),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewCancelProvideInsuranceCmd(), args)
}

func ExecMsgWithdrawInsurance(clientCtx client.Context, from string, insuranceId uint64, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		fmt.Sprintf("%d", insuranceId),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewWithdrawInsuranceCmd(), args)
}

func ExecMsgWithdrawInsuranceCommission(clientCtx client.Context, from string, insuranceId uint64, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		fmt.Sprintf("%d", insuranceId),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewWithdrawInsuranceCommissionCmd(), args)
}

func ExecMsgDepositInsurance(clientCtx client.Context, from string, insuranceId uint64, amount sdk.Coin, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		fmt.Sprintf("%d", insuranceId),
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewDepositInsuranceCmd(), args)
}

func ExecMsgLiquidStake(clientCtx client.Context, from string, amount sdk.Coin, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewLiquidStakeCmd(), args)
}

func ExecMsgLiquidUnstake(clientCtx client.Context, from string, amount sdk.Coin, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewLiquidUnstakeCmd(), args)
}

func ExecMsgSendCoins(clientCtx client.Context, from, to string, amount sdk.Coins, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		from,
		to,
		amount.String(),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, bankcli.NewSendTxCmd(), args)
}
