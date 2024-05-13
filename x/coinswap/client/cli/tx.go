package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetAddLiquidityCmd(),
		GetRemoveLiquidityCmd(),
		GetSwapCmd(),
	)

	return cmd
}

func GetAddLiquidityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity [max-coin] [standard-coin-amount] [minimum-liquidity] [duration]",
		Args:  cobra.ExactArgs(4),
		Short: "Add liquidity to a pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			depositCoin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid coins: %w", err)
			}

			standardCoinAmt, ok := sdkmath.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid standard coin amount: %s", args[1])
			}

			minLiquidity, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid minimum liquidity: %s", args[2])
			}

			duration, err := time.ParseDuration(args[3])
			if err != nil {
				return fmt.Errorf("invalid duration: %s", err)
			}

			deadline := time.Now().Add(duration)

			msg := types.NewMsgAddLiquidity(depositCoin, standardCoinAmt, minLiquidity, deadline.Unix(), clientCtx.GetFromAddress().String())

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetRemoveLiquidityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [min output coin amount] [liquidity coin to withdraw] [min output standard coin amount] [duration]",
		Args:  cobra.ExactArgs(4),
		Short: "Remove liquidity from a pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			minOutputCoin, ok := sdkmath.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("invalid output coin amount: %s", args[1])
			}

			liquidityCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid liquidity coin: %w", err)
			}

			minOutputStandardCoin, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid output standard coin amount: %s", args[2])
			}

			duration, err := time.ParseDuration(args[3])
			if err != nil {
				return fmt.Errorf("invalid duration: %s", err)
			}

			deadline := time.Now().Add(duration)

			msg := types.NewMsgRemoveLiquidity(minOutputCoin, liquidityCoin, minOutputStandardCoin, deadline.Unix(), clientCtx.GetFromAddress().String())

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetSwapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [input coin] [output coin] [isBuyOrder] [duration]",
		Args:  cobra.ExactArgs(4),
		Short: "Remove liquidity from a pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			inputCoin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid input coin: %w", err)
			}

			outputCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid output coin: %w", err)
			}

			isBuyOrder, err := strconv.ParseBool(args[2])
			if err != nil {
				return fmt.Errorf("invalid isBuyOrder value: %s", args[2])
			}

			duration, err := time.ParseDuration(args[3])
			if err != nil {
				return fmt.Errorf("invalid duration: %s", err)
			}

			deadline := time.Now().Add(duration)

			msg := types.NewMsgSwapOrder(
				types.Input{Address: clientCtx.GetFromAddress().String(), Coin: inputCoin},
				types.Output{Address: clientCtx.GetFromAddress().String(), Coin: outputCoin},
				deadline.Unix(),
				isBuyOrder,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
