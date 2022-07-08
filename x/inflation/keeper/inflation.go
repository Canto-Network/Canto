package keeper

import (

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto-Testnet-v2/v1/x/inflation/types"
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
	//remove minting coins to the incentives module

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
	distribution sdk.Dec,
) sdk.Coin {
	return sdk.NewCoin(
		coin.Denom,
		coin.Amount.ToDec().Mul(distribution).TruncateInt(),
	)
}

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	stakeSupply := k.stakingKeeper.StakingTokenSupply(ctx)

	if !stakeSupply.IsPositive() {
		return sdk.ZeroDec()
	}

	return k.stakingKeeper.TotalBondedTokens(ctx).ToDec().QuoInt(stakeSupply)
}

// GetCirculatingSupply returns the bank supply of the total inflation
func (k Keeper) GetCirculatingSupply(ctx sdk.Context) sdk.Dec {
	mintDenom := k.GetParams(ctx).MintDenom

	circulatingSupply := k.bankKeeper.GetSupply(ctx, mintDenom).Amount.ToDec()

	return circulatingSupply
}

//Set the curInflation as an Amino Marshalled object
func (k Keeper) SetCurInflation(ctx sdk.Context, curInflation sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)
	marshalledCurInflation, err := curInflation.MarshalAmino()
	if err != nil {
		return err
	}

	store.Set(types.KeyPrefixCurInflation, marshalledCurInflation)
	return nil
}

func (k Keeper) GetCurInflation(ctx sdk.Context) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixCurInflation)
	if len(bz) == 0 {
		return sdk.Dec{}, nil
	}

	var dec sdk.Dec
	if err := dec.UnmarshalAmino(bz); err != nil {
		return sdk.Dec{}, err
	}

	return dec, nil
}

//newInflation = min(maxInflation, Max(minInflation, curInflation * ((1 + (target - actual)) * adjustSpeed)
func (k Keeper) GetInflationRate(ctx sdk.Context) (sdk.Dec, error) {
	epp := k.GetEpochsPerPeriod(ctx)
	if epp == 0 {
		return sdk.ZeroDec(), nil
	}
	
	params := k.GetParams(ctx)
	//parameters for inflation calculation
	minInflation := params.ExponentialCalculation.MinInflation
	maxInflation := params.ExponentialCalculation.MaxInflation
	bondedTarget := params.ExponentialCalculation.BondingTarget
	adjustSpeed := params.ExponentialCalculation.AdjustSpeed
	curInflation, err := k.GetCurInflation(ctx)

	if err != nil {
		return sdk.Dec{}, err
	}

	//bondDifference
	curBonded := k.BondedRatio(ctx)
	bondDiff := bondedTarget.Sub(curBonded)

	//inflation annualized
	inflation := curInflation.Mul(adjustSpeed).Mul(bondDiff.Add(sdk.OneDec()))

	if inflation.LT(minInflation) {
		inflation = minInflation
	}
	if inflation.GT(maxInflation) {
		inflation = maxInflation
	}

	if err := k.SetCurInflation(ctx, inflation); err != nil {
		return sdk.Dec{}, err
	}
	//periodized inflation 
	return inflation.Quo(sdk.NewDec(epp)), nil
}

//requires that inflation has already been calculated
func (k Keeper) CalculateEpochMintProvision(ctx sdk.Context) (sdk.Dec, error) {
	if epp := k.GetEpochsPerPeriod(ctx);  epp == 0 {
		return sdk.ZeroDec(), nil
	}
	
	denomMint := k.GetParams(ctx).MintDenom
	//get the current circulatingSupply
	totalCirculatingSupply := k.GetCirculatingSupply(ctx)
	//distrKeeper supply of acanto is not counted
	feePool := k.distrKeeper.GetFeePool(ctx)
	circulatingSupply := totalCirculatingSupply.Sub(feePool.CommunityPool.AmountOf(denomMint))
	
	curInfl, err := k.GetCurInflation(ctx)
	if err != nil {
		return sdk.Dec{}, err
	}

	periodProvision := curInfl.Mul(circulatingSupply)
	return periodProvision, nil
}
