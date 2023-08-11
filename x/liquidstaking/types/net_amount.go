package types

import (
	"fmt"
)

func (nas NetAmountState) String() string {
	// Print all fields with field name
	return fmt.Sprintf(`NetAmountState:
	MintRate:                   %s
	LsTokensTotalSupply:        %s
	NetAmount: 	                %s	
	TotalLiquidTokens:          %s	
	RewardModuleAccountBalance: %s
	FeeRate:                    %s
	UtilizationRatio:           %s
	RemainingChunkSlots:        %s
	DiscountRate:               %s
	NumPairedChunks:            %s
	ChunkSize:                  %s
	TotalDelShares:             %s
	TotalRemainingRewards:      %s	
	TotalChunksBalance:         %s	
	TotalUnbondingBalance:      %s
	TotalInsuranceTokens:       %s
	TotalPairedInsuranceTokens: %s
    TotalUnpairingInsuranceTokens: %s
    TotalRemainingInsuranceCommissions: %s
`,
		nas.MintRate,
		nas.LsTokensTotalSupply,
		nas.NetAmount,
		nas.TotalLiquidTokens,
		nas.RewardModuleAccBalance,
		nas.FeeRate,
		nas.UtilizationRatio,
		nas.RemainingChunkSlots,
		nas.DiscountRate,
		nas.NumPairedChunks,
		nas.ChunkSize,
		nas.TotalDelShares,
		nas.TotalRemainingRewards,
		nas.TotalChunksBalance,
		nas.TotalUnbondingChunksBalance,
		nas.TotalInsuranceTokens,
		nas.TotalPairedInsuranceTokens,
		nas.TotalUnpairingInsuranceTokens,
		nas.TotalRemainingInsuranceCommissions,
	)
}
