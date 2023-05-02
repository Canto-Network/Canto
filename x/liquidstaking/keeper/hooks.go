package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CollectReward collects reward of chunk and paired insurance
// 1. Give commission based on chunk reward
// 2. Send rest of rewards to reward module account
func (k Keeper) CollectReward(ctx sdk.Context, chunk types.Chunk) {
	pairedInsurance, found := k.GetInsurance(ctx, chunk.PairedInsuranceId)
	if !found {
		panic(types.ErrNotFoundInsurance.Error())
	}

	bondDenom := k.stakingKeeper.BondDenom(ctx)
	chunkBalance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom)
	insuranceFee := chunkBalance.Amount.ToDec().Mul(pairedInsurance.FeeRate).TruncateInt()

	// Send pairedInsurance fee to the pairedInsurance fee pool
	if err := k.bankKeeper.SendCoins(
		ctx,
		chunk.DerivedAddress(),
		pairedInsurance.FeePoolAddress(),
		sdk.NewCoins(sdk.NewCoin(bondDenom, insuranceFee)),
	); err != nil {
		panic(err)
	}

	remained := chunkBalance.Amount.Sub(insuranceFee)
	if err := k.bankKeeper.SendCoins(
		ctx,
		chunk.DerivedAddress(),
		types.RewardPool,
		sdk.NewCoins(sdk.NewCoin(bondDenom, remained)),
	); err != nil {
		panic(err)
	}
}

func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	chunk, found := k.GetChunkByDerivedAddress(ctx, delAddr.String())
	if !found {
		return
	}
	k.CollectReward(ctx, chunk)
}

func (k Keeper) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	chunk, found := k.GetChunkByDerivedAddress(ctx, delAddr.String())
	if !found {
		return
	}
	k.CollectReward(ctx, chunk)
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
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.BeforeDelegationRemoved(ctx, delAddr, valAddr)
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.AfterDelegationModified(ctx, delAddr, valAddr)
}
func (h Hooks) BeforeValidatorSlashed(_ sdk.Context, _ sdk.ValAddress, _ sdk.Dec) {}
