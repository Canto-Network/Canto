package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ulule/deepcopier"
)

func (k Keeper) GetNetAmountState(ctx sdk.Context) types.NetAmountState {
	nase := k.GetNetAmountStateEssentials(ctx)
	nas := &types.NetAmountState{}
	deepcopier.Copy(&nase).To(nas)

	// fill insurance state
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	totalPairedInsuranceTokens, totalUnpairingInsuranceTokens, totalInsuranceTokens := sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	totalRemainingInsuranceCommissions := sdk.ZeroDec()
	k.IterateAllInsurances(ctx, func(insurance types.Insurance) (stop bool) {
		insuranceBalance := k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), bondDenom)
		commission := k.bankKeeper.GetBalance(ctx, insurance.FeePoolAddress(), bondDenom)
		switch insurance.Status {
		case types.INSURANCE_STATUS_PAIRED:
			totalPairedInsuranceTokens = totalPairedInsuranceTokens.Add(insuranceBalance.Amount)
		case types.INSURANCE_STATUS_UNPAIRING:
			totalUnpairingInsuranceTokens = totalUnpairingInsuranceTokens.Add(insuranceBalance.Amount)
		}
		totalInsuranceTokens = totalInsuranceTokens.Add(insuranceBalance.Amount)
		totalRemainingInsuranceCommissions = totalRemainingInsuranceCommissions.Add(commission.Amount.ToDec())
		return false
	})
	nas.TotalPairedInsuranceTokens = totalPairedInsuranceTokens
	nas.TotalUnpairingInsuranceTokens = totalUnpairingInsuranceTokens
	nas.TotalInsuranceTokens = totalInsuranceTokens
	nas.TotalRemainingInsuranceCommissions = totalRemainingInsuranceCommissions
	return *nas
}
