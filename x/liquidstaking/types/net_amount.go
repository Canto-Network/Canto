package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (nas NetAmountState) CalcNetAmount(rewardPoolBalance sdk.Int) sdk.Dec {
	return rewardPoolBalance.Add(nas.TotalChunksBalance).
		Add(nas.TotalLiquidTokens).
		Add(nas.TotalUnbondingBalance).ToDec().
		Add(nas.TotalRemainingRewards)
}

func (nas NetAmountState) CalcMintRate() sdk.Dec {
	if nas.NetAmount.IsNil() || !nas.NetAmount.IsPositive() {
		return sdk.ZeroDec()
	}
	return nas.LsTokensTotalSupply.ToDec().QuoTruncate(nas.NetAmount)
}

func (nas NetAmountState) Equal(nas2 NetAmountState) bool {
	return nas.LsTokensTotalSupply.Equal(nas2.LsTokensTotalSupply) &&
		nas.TotalChunksBalance.Equal(nas2.TotalChunksBalance) &&
		nas.TotalDelShares.Equal(nas2.TotalDelShares) &&
		nas.TotalRemainingRewards.Equal(nas2.TotalRemainingRewards) &&
		nas.TotalRemainingInsuranceCommissions.Equal(nas2.TotalRemainingInsuranceCommissions) &&
		nas.TotalLiquidTokens.Equal(nas2.TotalLiquidTokens) &&
		nas.TotalInsuranceTokens.Equal(nas2.TotalInsuranceTokens) &&
		nas.TotalInsuranceCommissions.Equal(nas2.TotalInsuranceCommissions) &&
		nas.TotalPairedInsuranceTokens.Equal(nas2.TotalPairedInsuranceTokens) &&
		nas.TotalPairedInsuranceCommissions.Equal(nas2.TotalPairedInsuranceCommissions) &&
		nas.TotalUnpairingInsuranceTokens.Equal(nas2.TotalUnpairingInsuranceTokens) &&
		nas.TotalUnpairingInsuranceCommissions.Equal(nas2.TotalUnpairingInsuranceCommissions) &&
		nas.TotalUnpairedInsuranceTokens.Equal(nas2.TotalUnpairedInsuranceTokens) &&
		nas.TotalUnpairedInsuranceCommissions.Equal(nas2.TotalUnpairedInsuranceCommissions) &&
		nas.TotalUnbondingBalance.Equal(nas2.TotalUnbondingBalance) &&
		nas.NetAmount.Equal(nas2.NetAmount) &&
		nas.MintRate.Equal(nas2.MintRate) &&
		nas.RewardModuleAccBalance.Equal(nas2.RewardModuleAccBalance)
}

func (nas NetAmountState) IsZeroState() bool {
	return nas.LsTokensTotalSupply.IsZero() &&
		nas.TotalChunksBalance.IsZero() &&
		nas.TotalDelShares.IsZero() &&
		nas.TotalRemainingRewards.IsZero() &&
		nas.TotalLiquidTokens.IsZero() &&
		// Currently total insurances includes Pairing insurances, so we should skip this
		//nas.TotalInsuranceTokens.IsZero() &&
		//nas.TotalInsuranceCommissions.IsZero() &&
		nas.TotalPairedInsuranceTokens.IsZero() &&
		nas.TotalPairedInsuranceCommissions.IsZero() &&
		nas.TotalUnpairingInsuranceTokens.IsZero() &&
		nas.TotalUnpairingInsuranceCommissions.IsZero() &&
		nas.TotalUnpairedInsuranceTokens.IsZero() &&
		nas.TotalUnpairedInsuranceCommissions.IsZero() &&
		nas.TotalUnbondingBalance.IsZero() &&
		nas.NetAmount.IsZero() &&
		nas.MintRate.IsZero() &&
		nas.RewardModuleAccBalance.IsZero()
}

func (nas NetAmountState) String() string {
	// Print all fields with field name
	return fmt.Sprintf(`NetAmountState:
	  LsTokensTotalSupply:   %s
	  TotalChunksBalance:    %s	
	  TotalDelShares:        %s
	  TotalRemainingRewards: %s	
      TotalRemainingInsuranceCommissions: %s
	  TotalLiquidTokens:     %s	
	  TotalInsuranceTokens:  %s
	  TotalInsuranceCommissons: %s
	  TotalPairedInsuranceTokens: %s
	  TotalPairedInsuranceCommissons: %s
      TotalUnpairingInsuranceTokens: %s
      TotalUnpairingInsuranceCommissons: %s
	  TotalUnpairedInsuranceTokens: %s
	  TotalUnpairedInsuranceCommissons: %s
	  TotalUnbondingBalance: %s
	  NetAmount:             %s
	  MintRate:              %s
	  RewardModuleAccountBalance: %s`,
		nas.LsTokensTotalSupply,
		nas.TotalChunksBalance,
		nas.TotalDelShares,
		nas.TotalRemainingRewards,
		nas.TotalRemainingInsuranceCommissions,
		nas.TotalLiquidTokens,
		nas.TotalInsuranceTokens,
		nas.TotalInsuranceCommissions,
		nas.TotalPairedInsuranceTokens,
		nas.TotalPairedInsuranceCommissions,
		nas.TotalUnpairingInsuranceTokens,
		nas.TotalUnpairingInsuranceCommissions,
		nas.TotalUnpairedInsuranceTokens,
		nas.TotalUnpairedInsuranceCommissions,
		nas.TotalUnbondingBalance,
		nas.NetAmount,
		nas.MintRate,
		nas.RewardModuleAccBalance)
}
