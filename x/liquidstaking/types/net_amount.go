package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (nas NetAmountState) CalcNetAmount(rewardPoolBalance sdk.Int) sdk.Dec {
	return rewardPoolBalance.Add(nas.TotalChunksBalance).
		Add(nas.TotalLiquidTokens).
		Add(nas.TotalUnbondingChunksBalance).ToDec().
		Add(nas.TotalRemainingRewards)
}

func (nas NetAmountState) CalcMintRate() sdk.Dec {
	if nas.NetAmount.IsNil() || !nas.NetAmount.IsPositive() {
		return sdk.ZeroDec()
	}
	return nas.LsTokensTotalSupply.ToDec().QuoTruncate(nas.NetAmount)
}

func (nas NetAmountState) Equal(nas2 NetAmountState) bool {
	return nas.MintRate.Equal(nas2.MintRate) &&
		nas.LsTokensTotalSupply.Equal(nas2.LsTokensTotalSupply) &&
		nas.NetAmount.Equal(nas2.NetAmount) &&
		nas.TotalLiquidTokens.Equal(nas2.TotalLiquidTokens) &&
		nas.RewardModuleAccBalance.Equal(nas2.RewardModuleAccBalance) &&
		nas.FeeRate.Equal(nas2.FeeRate) &&
		nas.UtilizationRatio.Equal(nas2.UtilizationRatio) &&
		nas.RemainingChunkSlots.Equal(nas2.RemainingChunkSlots) &&
		nas.DiscountRate.Equal(nas2.DiscountRate) &&
		nas.TotalDelShares.Equal(nas2.TotalDelShares) &&
		nas.TotalRemainingRewards.Equal(nas2.TotalRemainingRewards) &&
		nas.TotalChunksBalance.Equal(nas2.TotalChunksBalance) &&
		nas.TotalUnbondingChunksBalance.Equal(nas2.TotalUnbondingChunksBalance) &&
		nas.TotalInsuranceTokens.Equal(nas2.TotalInsuranceTokens) &&
		nas.TotalPairedInsuranceTokens.Equal(nas2.TotalPairedInsuranceTokens) &&
		nas.TotalUnpairingInsuranceTokens.Equal(nas2.TotalUnpairingInsuranceTokens) &&
		nas.TotalRemainingInsuranceCommissions.Equal(nas2.TotalRemainingInsuranceCommissions) &&
		nas.NumPairedChunks.Equal(nas2.NumPairedChunks)
	// Don't check ChunkSize because it is constant defined in module.
}

// IsZeroState checks if the NetAmountState is initial state or not.
// Some fields(e.g. TotalRemainingRewards) are not checked because they will rarely be zero.
func (nas NetAmountState) IsZeroState() bool {
	return nas.MintRate.IsZero() &&
		nas.LsTokensTotalSupply.IsZero() &&
		nas.NetAmount.IsZero() &&
		nas.TotalLiquidTokens.IsZero() &&
		nas.RewardModuleAccBalance.IsZero() &&
		nas.FeeRate.IsZero() &&
		nas.UtilizationRatio.IsZero() &&
		// As long as there is a total supply and a hard cap, this value will rarely be zero.
		// So we skip this
		// nas.RemainingChunkSlots.IsZero() &&
		nas.DiscountRate.IsZero() &&
		nas.TotalDelShares.IsZero() &&
		nas.TotalRemainingRewards.IsZero() &&
		nas.TotalChunksBalance.IsZero() &&
		nas.TotalUnbondingChunksBalance.IsZero() &&
		nas.NumPairedChunks.IsZero() &&
		// Don't check ChunkSize because it is constant defined in module.
		// nas.ChunkSize
		// Total insurances includes Pairing insurances, so we should skip this
		// nas.TotalInsuranceTokens.IsZero() &&
		nas.TotalPairedInsuranceTokens.IsZero() &&
		nas.TotalUnpairingInsuranceTokens.IsZero() &&
		nas.TotalRemainingInsuranceCommissions.IsZero()
}

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
