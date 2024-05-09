package csr

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cosmos/cosmos-sdk/version"

	csrv1 "github.com/Canto-Network/Canto/v7/api/canto/csr/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              csrv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current parameters of the CSR module",
				},
				{
					RpcMethod: "CSRs",
					Use:       "csrs",
					Short:     "Query all registered contracts and NFTs for the CSR module",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pagination"},
					},
				},
				{
					RpcMethod: "CSRByNFT",
					Use:       "nft [nftID]",
					Short:     "Query the CSR associated with a given NFT ID",
					Example:   fmt.Sprintf("%s query csr nft <address>", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "nftId"},
					},
				},
				{
					RpcMethod: "CSRByContract",
					Use:       "contract [address]",
					Short:     "Query the CSR associated with a given smart contract adddress",
					Example:   fmt.Sprintf("%s query csr contract <address>", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "Turnstile",
					Use:       "turnstile",
					Short:     "Query the address of the turnstile smart contract deployed by the module account",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              csrv1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true,
				},
			},
		},
	}
}
