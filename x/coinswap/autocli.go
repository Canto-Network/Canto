package coinswap

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cosmos/cosmos-sdk/version"

	coinswapv1 "github.com/Canto-Network/Canto/v7/api/canto/coinswap/v1"
	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              coinswapv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current coinswap parameters information",
					Example:   fmt.Sprintf("%s query %s params", version.AppName, types.ModuleName),
				},
				{
					RpcMethod: "LiquidityPool",
					Use:       "liquidity-pool [liquidity pool denom]",
					Short:     "Query proposals with optional filters",
					Example:   fmt.Sprintf("%s query %s liquidity-pool lpt-1", version.AppName, types.ModuleName),
				},
				{
					RpcMethod: "LiquidityPools",
					Use:       "liquidity-pools",
					Short:     "query all liquidity pools",
					Example:   fmt.Sprintf("%s query %s liquidity-pools", version.AppName, types.ModuleName),
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              coinswapv1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "AddLiquidity",
					Use:       "add-liquidity [max-coin] [standard-coin-amount] [minimum-liquidity] [duration]",
					Short:     "Add liquidity to a pool",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "max_token"},
						{ProtoField: "exact_standard_amt"},
						{ProtoField: "min_liquidity"},
						{ProtoField: "deadline"},
					},
				},
				{
					RpcMethod: "RemoveLiquidity",
					Use:       "remove-liquidity [min output coin amount] [liquidity coin to withdraw] [min output standard coin amount] [duration]",
					Short:     "Remove liquidity from a pair",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "withdraw_liquidity"},
						{ProtoField: "min_token"},
						{ProtoField: "min_standard_amt"},
						{ProtoField: "deadline"},
					},
				},
				{
					RpcMethod: "SwapCoin",
					Use:       "swap [input coin] [output coin] [isBuyOrder] [duration]",
					Short:     "Remove liquidity from a pair",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "input"},
						{ProtoField: "output"},
						{ProtoField: "deadline"},
						{ProtoField: "is_buy_order"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true,
				},
			},
		},
	}
}
