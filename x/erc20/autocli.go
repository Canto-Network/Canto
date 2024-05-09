package erc20

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	erc20v1 "github.com/Canto-Network/Canto/v7/api/canto/erc20/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              erc20v1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "TokenPairs",
					Use:       "token-pairs",
					Short:     "Gets registered token pairs",
				},
				{
					RpcMethod: "TokenPair",
					Use:       "token-pair [token]",
					Short:     "Get a registered token pair",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "token"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Gets erc20 params",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              erc20v1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "ConvertCoin",
					Use:       "convert-coin [coin] [receiver_hex]",
					Short:     "Convert a Cosmos coin to ERC20. When the receiver [optional] is omitted, the ERC20 tokens are transferred to the sender.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "coin"},
						{ProtoField: "receiver"},
						{ProtoField: "sender"},
					},
				},
				{
					RpcMethod: "ConvertERC20",
					Use:       "convert-erc20 [contract-address] [amount] [receiver]",
					Short:     "Convert an ERC20 token to Cosmos coin.  When the receiver [optional] is omitted, the Cosmos coins are transferred to the sender.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
						{ProtoField: "amount"},
						{ProtoField: "receiver"},
						{ProtoField: "sender"},
					},
				},
				{
					RpcMethod: "RegisterCoinProposal",
					Skip:      true,
				},
				{
					RpcMethod: "RegisterERC20Proposal",
					Skip:      true,
				},
				{
					RpcMethod: "ToggleTokenConversionProposal",
					Skip:      true,
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true,
				},
			},
		},
	}
}
