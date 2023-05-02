package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (nas NetAmountState) CalcNetAmount(rewardPoolBalance sdk.Int) sdk.Dec {
	// TODO: Add reward module account's balance
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

func (nas NetAmountState) String() string {
	// Print all fields with field name
	return fmt.Sprintf(`NetAmountState:
	  LsTokensTotalSupply:   %s
	  TotalChunksBalance:    %s	
	  TotalDelShares:        %s
	  TotalRemainingRewards: %s	
	  TotalLiquidTokens:     %s	
	  TotalInsuranceTokens:  %s
	  TotalUnbondingBalance: %s
	  NetAmount:             %s
	  MintRate:              %s`,
		nas.LsTokensTotalSupply,
		nas.TotalChunksBalance,
		nas.TotalDelShares,
		nas.TotalRemainingRewards,
		nas.TotalLiquidTokens,
		nas.TotalInsuranceTokens,
		nas.TotalUnbondingBalance,
		nas.NetAmount,
		nas.MintRate)
}
