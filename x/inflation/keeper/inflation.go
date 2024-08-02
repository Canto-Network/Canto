package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v7/x/inflation/types"
)

// MintAndAllocateInflation performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(
	ctx sdk.Context,
	coin sdk.Coin,
) (
	staking, communityPool sdk.Coins,
	err error,
) {
	// Mint coins for distribution
	if err := k.MintCoins(ctx, coin); err != nil {
		return nil, nil, err
	}

	// Allocate minted coins according to allocation proportions (staking, usage
	// incentives, community pool)
	return k.AllocateExponentialInflation(ctx, coin)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, coin sdk.Coin) error {
	coins := sdk.NewCoins(coin)

	// skip as no coins need to be minted
	if coins.Empty() {
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
}

// AllocateExponentialInflation allocates coins from the inflation to external
// modules according to allocation proportions:
//   - staking rewards -> sdk `auth` module fee collector
//   - community pool -> `sdk `distr` module community pool
func (k Keeper) AllocateExponentialInflation(
	ctx sdk.Context,
	mintedCoin sdk.Coin,
) (
	staking, communityPool sdk.Coins,
	err error,
) {
	params := k.GetParams(ctx)
	proportions := params.InflationDistribution
	// Allocate staking rewards into fee collector account
	staking = sdk.NewCoins(k.GetProportions(ctx, mintedCoin, proportions.StakingRewards))
	err = k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		k.feeCollectorName,
		staking,
	)
	if err != nil {
		return nil, nil, err
	}

	// Allocate community pool amount (remaining module balance) to community
	// pool address
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	communityPool = k.bankKeeper.GetAllBalances(ctx, moduleAddr)

	err = k.distrKeeper.FundCommunityPool(
		ctx,
		communityPool,
		moduleAddr,
	)

	if err != nil {
		return nil, nil, err
	}

	return staking, communityPool, nil
}

// GetAllocationProportion calculates the proportion of coins that is to be
// allocated during inflation for a given distribution.
func (k Keeper) GetProportions(
	ctx sdk.Context,
	coin sdk.Coin,
	distribution sdkmath.LegacyDec,
) sdk.Coin {
	return sdk.NewCoin(
		coin.Denom,
		coin.Amount.ToLegacyDec().Mul(distribution).TruncateInt(),
	)
}

// BondedRatio the fraction of the staking tokens which are currently bonded
// It doesn't consider team allocation for inflation
func (k Keeper) BondedRatio(ctx sdk.Context) sdkmath.LegacyDec {
	stakeSupply, err := k.stakingKeeper.StakingTokenSupply(ctx)

	if err != nil || !stakeSupply.IsPositive() {
		return sdkmath.LegacyZeroDec()
	}

	totalBonded, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return sdkmath.LegacyZeroDec()
	}

	return totalBonded.ToLegacyDec().QuoInt(stakeSupply)
}

// GetCirculatingSupply returns the bank supply of the mintDenom
func (k Keeper) GetCirculatingSupply(ctx sdk.Context) sdkmath.LegacyDec {
	mintDenom := k.GetParams(ctx).MintDenom

	circulatingSupply := k.bankKeeper.GetSupply(ctx, mintDenom).Amount.ToLegacyDec()

	return circulatingSupply
}

// GetInflationRate returns the inflation rate for the current period.
func (k Keeper) GetInflationRate(ctx sdk.Context) sdkmath.LegacyDec {
	epochMintProvision, _ := k.GetEpochMintProvision(ctx)
	if epochMintProvision.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	epp := k.GetEpochsPerPeriod(ctx)
	if epp == 0 {
		return sdkmath.LegacyZeroDec()
	}

	epochsPerPeriod := sdkmath.LegacyNewDec(epp)

	circulatingSupply := k.GetCirculatingSupply(ctx)
	if circulatingSupply.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	// EpochMintProvision * 365 / circulatingSupply * 100
	return epochMintProvision.Mul(epochsPerPeriod).Quo(circulatingSupply).Mul(sdkmath.LegacyNewDec(100))
}
