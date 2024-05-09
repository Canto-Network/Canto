package onboarding

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	onboardingv1 "github.com/Canto-Network/Canto/v7/api/canto/onboarding/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              onboardingv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Gets onboarding params",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			EnhanceCustomCommand: false,
		},
	}
}
