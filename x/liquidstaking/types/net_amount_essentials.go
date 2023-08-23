package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (nase NetAmountStateEssentials) CalcNetAmount() sdk.Dec {
	return nase.RewardModuleAccBalance.Add(nase.TotalChunksBalance).
		Add(nase.TotalLiquidTokens).
		Add(nase.TotalUnbondingChunksBalance).ToDec().
		Add(nase.TotalRemainingRewards)
}

func (nase NetAmountStateEssentials) CalcMintRate() sdk.Dec {
	if nase.NetAmount.IsNil() || !nase.NetAmount.IsPositive() {
		return sdk.ZeroDec()
	}
	return nase.LsTokensTotalSupply.ToDec().QuoTruncate(nase.NetAmount)
}

// CalcDiscountRate calculates the current discount rate.
// reward module account's balance / (num paired chunks * chunk size)
func (nase NetAmountStateEssentials) CalcDiscountRate(maximumDiscountRate sdk.Dec) sdk.Dec {
	if nase.RewardModuleAccBalance.IsZero() || maximumDiscountRate.IsZero() || !nase.NetAmount.IsPositive() {
		return sdk.ZeroDec()
	}
	discountRate := nase.RewardModuleAccBalance.ToDec().QuoTruncate(nase.NetAmount)
	return sdk.MinDec(discountRate, sdk.MinDec(MaximumDiscountRateCap, maximumDiscountRate))
}

func (nase NetAmountStateEssentials) Equal(nase2 NetAmountStateEssentials) bool {
	return nase.MintRate.Equal(nase2.MintRate) &&
		nase.LsTokensTotalSupply.Equal(nase2.LsTokensTotalSupply) &&
		nase.NetAmount.Equal(nase2.NetAmount) &&
		nase.TotalLiquidTokens.Equal(nase2.TotalLiquidTokens) &&
		nase.RewardModuleAccBalance.Equal(nase2.RewardModuleAccBalance) &&
		nase.FeeRate.Equal(nase2.FeeRate) &&
		nase.UtilizationRatio.Equal(nase2.UtilizationRatio) &&
		nase.RemainingChunkSlots.Equal(nase2.RemainingChunkSlots) &&
		nase.DiscountRate.Equal(nase2.DiscountRate) &&
		nase.TotalDelShares.Equal(nase2.TotalDelShares) &&
		nase.TotalRemainingRewards.Equal(nase2.TotalRemainingRewards) &&
		nase.TotalChunksBalance.Equal(nase2.TotalChunksBalance) &&
		nase.TotalUnbondingChunksBalance.Equal(nase2.TotalUnbondingChunksBalance) &&
		nase.NumPairedChunks.Equal(nase2.NumPairedChunks)
	// Don't check ChunkSize because it is constant defined in module.
}

// IsZeroState checks if the NetAmountState is initial state or not.
// Some fields(e.g. TotalRemainingRewards) are not checked because they will rarely be zero.
func (nase NetAmountStateEssentials) IsZeroState() bool {
	return nase.MintRate.IsZero() &&
		nase.LsTokensTotalSupply.IsZero() &&
		nase.NetAmount.IsZero() &&
		nase.TotalLiquidTokens.IsZero() &&
		nase.RewardModuleAccBalance.IsZero() &&
		nase.FeeRate.IsZero() &&
		nase.UtilizationRatio.IsZero() &&
		// As long as there is a total supply and a hard cap, this value will rarely be zero.
		// So we skip this
		// nase.RemainingChunkSlots.IsZero() &&
		nase.DiscountRate.IsZero() &&
		nase.TotalDelShares.IsZero() &&
		nase.TotalRemainingRewards.IsZero() &&
		nase.TotalChunksBalance.IsZero() &&
		nase.TotalUnbondingChunksBalance.IsZero() &&
		// Don't check ChunkSize because it is constant defined in module.
		// nase.ChunkSize
		nase.NumPairedChunks.IsZero()
}

func (nase NetAmountStateEssentials) String() string {
	// Print all fields with field name
	return fmt.Sprintf(`NetAmountStateEssentials:
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
`,
		nase.MintRate,
		nase.LsTokensTotalSupply,
		nase.NetAmount,
		nase.TotalLiquidTokens,
		nase.RewardModuleAccBalance,
		nase.FeeRate,
		nase.UtilizationRatio,
		nase.RemainingChunkSlots,
		nase.DiscountRate,
		nase.NumPairedChunks,
		nase.ChunkSize,
		nase.TotalDelShares,
		nase.TotalRemainingRewards,
		nase.TotalChunksBalance,
		nase.TotalUnbondingChunksBalance,
	)
}
