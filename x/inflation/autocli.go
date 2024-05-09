package inflation

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	inflationv1 "github.com/Canto-Network/Canto/v7/api/canto/inflation/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              inflationv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Period",
					Use:       "period",
					Short:     "Query the current inflation period",
				},
				{
					RpcMethod: "EpochMintProvision",
					Use:       "epoch-mint-provision",
					Short:     "Query the current inflation epoch provisions value",
				},
				{
					RpcMethod: "SkippedEpochs",
					Use:       "skipped-epochs",
					Short:     "Query the current number of skipped epochs",
				},
				{
					RpcMethod: "CirculatingSupply",
					Use:       "circulating-supply",
					Short:     "Query the current supply of tokens in circulation",
				},
				{
					RpcMethod: "InflationRate",
					Use:       "inflation-rate",
					Short:     "Query the inflation rate of the current period",
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current inflation parameters",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			EnhanceCustomCommand: false,
		},
	}
}
