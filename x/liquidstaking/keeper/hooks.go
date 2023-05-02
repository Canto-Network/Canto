package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements StakingHooks interface
func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// TODO: Fee distribution
	// 1. Reward will be distributed to the delegator address (= derived chunk address)
	// 2. Send (balance of derived chunk address) x insurance.feeRate to the insurance fee pool
	chunk, found := k.GetChunkByDerivedAddress(ctx, delAddr.String())
	if !found {
		return
	}
	insurance, found := k.GetInsurance(ctx, chunk.InsuranceId)
	if !found {
		panic("insurance not found")
	}
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	chunkBalance := k.bankKeeper.GetBalance(ctx, delAddr, bondDenom)
	insuranceFee := chunkBalance.Amount.ToDec().Mul(insurance.FeeRate).TruncateInt()
	// Send insurance fee to the insurance fee pool
	if err := k.bankKeeper.SendCoins(ctx, delAddr, insurance.FeePoolAddress(), sdk.NewCoins(sdk.NewCoin(bondDenom, insuranceFee))); err != nil {
		panic(err)
	}
	remained := chunkBalance.Amount.Sub(insuranceFee)
	if err := k.bankKeeper.SendCoins(ctx, delAddr, types.RewardPool, sdk.NewCoins(sdk.NewCoin(bondDenom, remained))); err != nil {
		panic(err)
	}
}

type Hooks struct {
	k Keeper
}

var _ types.StakingHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterValidatorBonded(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress)          {}
func (h Hooks) AfterValidatorRemoved(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress)         {}
func (h Hooks) AfterValidatorCreated(_ sdk.Context, _ sdk.ValAddress)                            {}
func (h Hooks) AfterValidatorBeginUnbonding(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress)  {}
func (h Hooks) BeforeValidatorModified(_ sdk.Context, _ sdk.ValAddress)                          {}
func (h Hooks) BeforeDelegationCreated(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress)        {}
func (h Hooks) BeforeDelegationSharesModified(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) {}
func (h Hooks) BeforeDelegationRemoved(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress)        {}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.AfterDelegationModified(ctx, delAddr, valAddr)
}
func (h Hooks) BeforeValidatorSlashed(_ sdk.Context, _ sdk.ValAddress, _ sdk.Dec) {}
