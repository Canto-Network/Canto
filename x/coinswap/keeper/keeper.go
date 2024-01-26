package keeper

import (
	"fmt"
	"strconv"

	gogotypes "github.com/gogo/protobuf/types"

	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

// Keeper of the coinswap store
type Keeper struct {
	cdc              codec.BinaryCodec
	storeService     store.KVStoreService
	bk               types.BankKeeper
	ak               types.AccountKeeper
	paramSpace       paramstypes.Subspace
	feeCollectorName string
	blockedAddrs     map[string]bool

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper returns a coinswap keeper. It handles:
// - creating new ModuleAccounts for each trading pair
// - burning and minting liquidity coins
// - sending to and from ModuleAccounts
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	paramSpace paramstypes.Subspace,
	bk types.BankKeeper,
	ak types.AccountKeeper,
	blockedAddrs map[string]bool,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure coinswap module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeService:     storeService,
		bk:               bk,
		ak:               ak,
		cdc:              cdc,
		paramSpace:       paramSpace,
		blockedAddrs:     blockedAddrs,
		feeCollectorName: feeCollectorName,
		authority:        authority,
	}
}

// GetAuthority returns the x/coinswap module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Swap execute swap order in specified pool
func (k Keeper) Swap(ctx sdk.Context, msg *types.MsgSwapOrder) error {
	var amount sdkmath.Int
	var err error

	standardDenom, err := k.GetStandardDenom(ctx)
	if err != nil {
		return err
	}

	isDoubleSwap := (msg.Input.Coin.Denom != standardDenom) && (msg.Output.Coin.Denom != standardDenom)
	if isDoubleSwap {
		return errorsmod.Wrapf(types.ErrNotContainStandardDenom, "unsupported swap: standard coin must be in either Input or Output")
	}
	if msg.IsBuyOrder {
		amount, err = k.TradeInputForExactOutput(ctx, msg.Input, msg.Output)
	} else {
		amount, err = k.TradeExactInputForOutput(ctx, msg.Input, msg.Output)
	}
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwap,
			sdk.NewAttribute(types.AttributeValueAmount, amount.String()),
			sdk.NewAttribute(types.AttributeValueSender, msg.Input.Address),
			sdk.NewAttribute(types.AttributeValueRecipient, msg.Output.Address),
			sdk.NewAttribute(types.AttributeValueIsBuyOrder, strconv.FormatBool(msg.IsBuyOrder)),
			sdk.NewAttribute(types.AttributeValueTokenPair, types.GetTokenPairByDenom(msg.Input.Coin.Denom, msg.Output.Coin.Denom)),
		),
	)

	return nil
}

// AddLiquidity adds liquidity to the specified pool
func (k Keeper) AddLiquidity(ctx sdk.Context, msg *types.MsgAddLiquidity) (sdk.Coin, error) {
	standardDenom, err := k.GetStandardDenom(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}
	if standardDenom == msg.MaxToken.Denom {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidDenom,
			"MaxToken: %s should not be StandardDenom", msg.MaxToken.String())
	}

	params := k.GetParams(ctx)
	if !params.MaxSwapAmount.AmountOf(msg.MaxToken.Denom).IsPositive() {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidDenom,
			"MaxToken %s is not registered in max swap amount", msg.MaxToken.Denom)
	}

	var mintLiquidityAmt sdkmath.Int
	var depositToken sdk.Coin
	var standardCoin = sdk.NewCoin(standardDenom, msg.ExactStandardAmt)

	poolId := types.GetPoolId(msg.MaxToken.Denom)
	pool, exists := k.GetPool(ctx, poolId)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdk.Coin{}, err
	}

	// calculate amount of UNI to be minted for sender
	// and coin amount to be deposited
	if !exists {
		// deduct the user's fee for creating a Liquidity pool
		if err := k.DeductPoolCreationFee(ctx, sender); err != nil {
			return sdk.Coin{}, err
		}

		mintLiquidityAmt = msg.ExactStandardAmt

		if mintLiquidityAmt.GT(params.MaxStandardCoinPerPool) {
			return sdk.Coin{}, errorsmod.Wrap(types.ErrMaxedStandardDenom, fmt.Sprintf("liquidity amount not met, max standard coin amount: no bigger than %s, actual: %s", params.MaxStandardCoinPerPool.String(), mintLiquidityAmt.String()))
		}

		if mintLiquidityAmt.LT(msg.MinLiquidity) {
			return sdk.Coin{}, errorsmod.Wrap(types.ErrConstraintNotMet, fmt.Sprintf("liquidity amount not met, user expected: no less than %s, actual: %s", msg.MinLiquidity.String(), mintLiquidityAmt.String()))
		}

		depositToken = sdk.NewCoin(msg.MaxToken.Denom, msg.MaxToken.Amount)
		pool = k.CreatePool(ctx, msg.MaxToken.Denom)
	} else {
		balances, err := k.GetPoolBalances(ctx, pool.EscrowAddress)
		if err != nil {
			return sdk.Coin{}, err
		}

		standardReserveAmt := balances.AmountOf(standardDenom)
		tokenReserveAmt := balances.AmountOf(msg.MaxToken.Denom)
		liquidity := k.bk.GetSupply(ctx, pool.LptDenom).Amount

		if liquidity.Equal(sdkmath.ZeroInt()) {
			// pool exists, but it is empty
			// same with initial liquidity provide
			mintLiquidityAmt = msg.ExactStandardAmt

			if mintLiquidityAmt.GT(params.MaxStandardCoinPerPool) {
				return sdk.Coin{}, errorsmod.Wrap(types.ErrMaxedStandardDenom, fmt.Sprintf("liquidity amount not met, max standard coin amount: no bigger than %s, actual: %s", params.MaxStandardCoinPerPool.String(), mintLiquidityAmt.String()))
			}

			if mintLiquidityAmt.LT(msg.MinLiquidity) {
				return sdk.Coin{}, errorsmod.Wrap(types.ErrConstraintNotMet, fmt.Sprintf("liquidity amount not met, user expected: no less than %s, actual: %s", msg.MinLiquidity.String(), mintLiquidityAmt.String()))
			}

			depositToken = sdk.NewCoin(msg.MaxToken.Denom, msg.MaxToken.Amount)

		} else {
			if standardReserveAmt.GTE(params.MaxStandardCoinPerPool) {
				return sdk.Coin{}, errorsmod.Wrap(types.ErrMaxedStandardDenom, fmt.Sprintf("pool standard coin is maxed out: %s", params.MaxStandardCoinPerPool.String()))
			}

			maxStandardInputAmt := sdkmath.MinInt(msg.ExactStandardAmt, params.MaxStandardCoinPerPool.Sub(standardReserveAmt))
			mintLiquidityAmt = (liquidity.Mul(maxStandardInputAmt)).Quo(standardReserveAmt)
			if mintLiquidityAmt.LT(msg.MinLiquidity) {
				return sdk.Coin{}, errorsmod.Wrap(types.ErrConstraintNotMet, fmt.Sprintf("liquidity amount not met, user expected: no less than %s, actual: %s", msg.MinLiquidity.String(), mintLiquidityAmt.String()))
			}

			depositAmt := (tokenReserveAmt.Mul(maxStandardInputAmt)).Quo(standardReserveAmt).AddRaw(1)
			depositToken = sdk.NewCoin(msg.MaxToken.Denom, depositAmt)
			standardCoin = sdk.NewCoin(standardDenom, maxStandardInputAmt)
			if depositAmt.GT(msg.MaxToken.Amount) {
				return sdk.Coin{}, errorsmod.Wrap(types.ErrConstraintNotMet, fmt.Sprintf("token amount not met, user expected: no more than %s, actual: %s", msg.MaxToken.String(), depositToken.String()))
			}
		}
	}

	reservePoolAddress, err := sdk.AccAddressFromBech32(pool.EscrowAddress)
	if err != nil {
		return sdk.Coin{}, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddLiquidity,
			sdk.NewAttribute(types.AttributeValueSender, msg.Sender),
			sdk.NewAttribute(types.AttributeValueTokenPair, types.GetTokenPairByDenom(msg.MaxToken.Denom, standardDenom)),
		),
	)
	return k.addLiquidity(ctx, sender, reservePoolAddress, standardCoin, depositToken, pool.LptDenom, mintLiquidityAmt)
}

func (k Keeper) addLiquidity(ctx sdk.Context,
	sender sdk.AccAddress,
	reservePoolAddress sdk.AccAddress,
	standardCoin, token sdk.Coin,
	lptDenom string,
	mintLiquidityAmt sdkmath.Int,
) (sdk.Coin, error) {
	depositedTokens := sdk.NewCoins(standardCoin, token)
	// transfer deposited token into coinswaps Account
	if err := k.bk.SendCoins(ctx, sender, reservePoolAddress, depositedTokens); err != nil {
		return sdk.Coin{}, err
	}

	mintToken := sdk.NewCoin(lptDenom, mintLiquidityAmt)
	mintTokens := sdk.NewCoins(mintToken)
	if err := k.bk.MintCoins(ctx, types.ModuleName, mintTokens); err != nil {
		return sdk.Coin{}, err
	}
	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, mintTokens); err != nil {
		return sdk.Coin{}, err
	}

	return mintToken, nil
}

// RemoveLiquidity removes liquidity from the specified pool
func (k Keeper) RemoveLiquidity(ctx sdk.Context, msg *types.MsgRemoveLiquidity) (sdk.Coins, error) {
	standardDenom, err := k.GetStandardDenom(ctx)
	if err != nil {
		return nil, err
	}

	pool, exists := k.GetPoolByLptDenom(ctx, msg.WithdrawLiquidity.Denom)
	if !exists {
		return nil, errorsmod.Wrapf(types.ErrReservePoolNotExists, "liquidity pool token: %s", msg.WithdrawLiquidity.Denom)
	}

	balances, err := k.GetPoolBalances(ctx, pool.EscrowAddress)
	if err != nil {
		return nil, err
	}

	lptDenom := msg.WithdrawLiquidity.Denom
	minTokenDenom := pool.CounterpartyDenom

	standardReserveAmt := balances.AmountOf(standardDenom)
	tokenReserveAmt := balances.AmountOf(minTokenDenom)
	liquidityReserve := k.bk.GetSupply(ctx, lptDenom).Amount
	if standardReserveAmt.LT(msg.MinStandardAmt) {
		return nil, errorsmod.Wrap(types.ErrInsufficientFunds, fmt.Sprintf("insufficient %s funds, user expected: %s, actual: %s", standardDenom, msg.MinStandardAmt.String(), standardReserveAmt.String()))
	}
	if tokenReserveAmt.LT(msg.MinToken) {
		return nil, errorsmod.Wrap(types.ErrInsufficientFunds, fmt.Sprintf("insufficient %s funds, user expected: %s, actual: %s", minTokenDenom, msg.MinToken.String(), tokenReserveAmt.String()))
	}
	if liquidityReserve.LT(msg.WithdrawLiquidity.Amount) {
		return nil, errorsmod.Wrap(types.ErrInsufficientFunds, fmt.Sprintf("insufficient %s funds, user expected: %s, actual: %s", lptDenom, msg.WithdrawLiquidity.Amount.String(), liquidityReserve.String()))
	}

	// calculate amount of UNI to be burned for sender
	// and coin amount to be returned
	standardWithdrawAmt := msg.WithdrawLiquidity.Amount.Mul(standardReserveAmt).Quo(liquidityReserve)
	tokenWithdrawnAmt := msg.WithdrawLiquidity.Amount.Mul(tokenReserveAmt).Quo(liquidityReserve)

	standardWithdrawCoin := sdk.NewCoin(standardDenom, standardWithdrawAmt)
	tokenWithdrawCoin := sdk.NewCoin(minTokenDenom, tokenWithdrawnAmt)
	deductUniCoin := msg.WithdrawLiquidity

	if standardWithdrawCoin.Amount.LT(msg.MinStandardAmt) {
		return nil, errorsmod.Wrap(types.ErrConstraintNotMet, fmt.Sprintf("iris amount not met, user expected: no less than %s, actual: %s", sdk.NewCoin(standardDenom, msg.MinStandardAmt).String(), standardWithdrawCoin.String()))
	}
	if tokenWithdrawCoin.Amount.LT(msg.MinToken) {
		return nil, errorsmod.Wrap(types.ErrConstraintNotMet, fmt.Sprintf("token amount not met, user expected: no less than %s, actual: %s", sdk.NewCoin(minTokenDenom, msg.MinToken).String(), tokenWithdrawCoin.String()))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveLiquidity,
			sdk.NewAttribute(types.AttributeValueSender, msg.Sender),
			sdk.NewAttribute(types.AttributeValueTokenPair, types.GetTokenPairByDenom(minTokenDenom, standardDenom)),
		),
	)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	poolAddr, err := sdk.AccAddressFromBech32(pool.EscrowAddress)
	if err != nil {
		return nil, err
	}

	return k.removeLiquidity(ctx, poolAddr, sender, deductUniCoin, standardWithdrawCoin, tokenWithdrawCoin)
}

func (k Keeper) removeLiquidity(ctx sdk.Context, poolAddr, sender sdk.AccAddress, deductUniCoin, standardWithdrawCoin, tokenWithdrawCoin sdk.Coin) (sdk.Coins, error) {
	deltaCoins := sdk.NewCoins(deductUniCoin)

	// send liquidity vouchers to be burned from sender account to module account
	if err := k.bk.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, deltaCoins); err != nil {
		return nil, err
	}
	// burn liquidity vouchers of reserve pool from module account
	if err := k.bk.BurnCoins(ctx, types.ModuleName, deltaCoins); err != nil {
		return nil, err
	}

	// transfer withdrawn liquidity from coinswap reserve pool account to sender account
	coins := sdk.NewCoins(standardWithdrawCoin, tokenWithdrawCoin)

	return coins, k.bk.SendCoins(ctx, poolAddr, sender, coins)
}

// GetParams gets the parameters for the coinswap module.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var swapParams types.Params
	k.paramSpace.GetParamSet(ctx, &swapParams)
	return swapParams
}

// SetParams sets the parameters for the coinswap module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// SetStandardDenom sets the standard denom for the coinswap module.
func (k Keeper) SetStandardDenom(ctx sdk.Context, denom string) error {
	store := k.storeService.OpenKVStore(ctx)
	denomWrap := gogotypes.StringValue{Value: denom}
	bz := k.cdc.MustMarshal(&denomWrap)
	err := store.Set(types.KeyStandardDenom, bz)
	if err != nil {
		return err
	}
	return nil
}

// GetStandardDenom returns the standard denom of the coinswap module.
func (k Keeper) GetStandardDenom(ctx sdk.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.KeyStandardDenom)
	if len(bz) == 0 {
		return "", err
	}

	var denomWrap = gogotypes.StringValue{}
	k.cdc.MustUnmarshal(bz, &denomWrap)
	return denomWrap.Value, nil
}
