package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
)

// GetTxCmd returns the transaction commands for the module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewLiquidStakeCmd(),
		NewLiquidUnstakeCmd(),
		NewProvideInsuranceCmd(),
		NewCancelProvideInsuranceCmd(),
		NewDepositInsuranceCmd(),
		NewWithdrawInsuranceCmd(),
		NewWithdrawInsuranceCommissionCmd(),
	)

	return cmd
}

func NewLiquidStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-stake [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "liquid stake",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquid-stake coin.
Example:
$ %s tx %s liquid-stake 5000000acanto --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgLiquidStake(clientCtx.GetFromAddress().String(), coin)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewLiquidUnstakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-unstake [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "liquid unstake",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquid-unstake coin.

Example:
$ %s tx %s liquid-unstake 5000000acanto --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgLiquidUnstake(clientCtx.GetFromAddress().String(), coin)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewProvideInsuranceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provide-insurance [validator-address] [amount] [fee-rate]",
		Args:  cobra.ExactArgs(3),
		Short: "insurance provide for chunk",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Provide insurance for chunk.

Example:
$ %s tx %s provide-insurance cantovaloper1gxl6usug4cz60yhpsjj7vw7vzysrz772yxjzsf 50acanto 0.01 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			val, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			feeRate, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgProvideInsurance(clientCtx.GetFromAddress().String(), val.String(), coin, feeRate)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCancelProvideInsuranceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-insurance-provide",
		Args:  cobra.ExactArgs(1),
		Short: "cancel insurance provide for chunk",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel-inusrance-provide for chunk.

Example:
$ %s tx %s cancel-insurance-provide 1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// arg must be converted to a uint
			insuranceId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelProvideInsurance(clientCtx.GetFromAddress().String(), insuranceId)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewDepositInsuranceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-insurance",
		Args:  cobra.ExactArgs(1),
		Short: "deposit more coin to insurance",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Deposit-inusrance.

Example:
$ %s tx %s deposit-insurance 2 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// arg must be converted to a uint
			insuranceId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgDepositInsurance(clientCtx.GetFromAddress().String(), insuranceId, coin)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewWithdrawInsuranceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-insurance",
		Args:  cobra.ExactArgs(1),
		Short: "withdraw insurance",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw-inusrance.

Example:
$ %s tx %s withdraw-insurance 1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// arg must be converted to a uint
			insuranceId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawInsurance(clientCtx.GetFromAddress().String(), insuranceId)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewWithdrawInsuranceCommissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-insurance-commission",
		Args:  cobra.ExactArgs(1),
		Short: "withdraw insurance commission",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw-inusrance.

Example:
$ %s tx %s withdraw-insurance 1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// arg must be converted to a uint
			insuranceId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawInsuranceCommission(clientCtx.GetFromAddress().String(), insuranceId)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
