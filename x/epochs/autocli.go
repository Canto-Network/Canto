package epochs

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cosmos/cosmos-sdk/version"

	epochsv1 "github.com/Canto-Network/Canto/v7/api/canto/epochs/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              epochsv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "EpochInfos",
					Use:       "epoch-infos",
					Short:     "Query running epochInfos",
					Example:   fmt.Sprintf("%s query epochs epoch-infos", version.AppName),
				},
				{
					RpcMethod: "CurrentEpoch",
					Use:       "current-epoch",
					Short:     "Query current epoch by specified identifier",
					Example:   fmt.Sprintf("%s query epochs current-epoch week", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "identifier"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			EnhanceCustomCommand: false,
		},
	}
}
