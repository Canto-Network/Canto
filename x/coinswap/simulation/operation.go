package simulation

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/Canto-Network/Canto/v8/app/params"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v8/x/coinswap/keeper"
	"github.com/Canto-Network/Canto/v8/x/coinswap/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgSwapOrder       = "op_weight_msg_swap_order"
	OpWeightMsgAddLiquidity    = "op_weight_msg_add_liquidity"
	OpWeightMsgRemoveLiquidity = "op_weight_msg_remove_liquidity"
)

var (
	TypeMsgSwapOrder       = sdk.MsgTypeURL(&types.MsgSwapOrder{})
	TypeMsgAddLiquidity    = sdk.MsgTypeURL(&types.MsgAddLiquidity{})
	TypeMsgRemoveLiquidity = sdk.MsgTypeURL(&types.MsgRemoveLiquidity{})
)

func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	k keeper.Keeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
) simulation.WeightedOperations {
	var (
		weightSwap   int
		weightAdd    int
		weightRemove int
	)

	appParams.GetOrGenerate(
		OpWeightMsgSwapOrder, &weightSwap, nil,
		func(_ *rand.Rand) {
			weightSwap = params.DefaultWeightMsgSwapOrder
		},
	)

	appParams.GetOrGenerate(
		OpWeightMsgAddLiquidity, &weightAdd, nil,
		func(_ *rand.Rand) {
			weightAdd = params.DefaultWeightMsgAddLiquidity
		},
	)

	appParams.GetOrGenerate(
		OpWeightMsgRemoveLiquidity, &weightRemove, nil,
		func(_ *rand.Rand) {
			weightRemove = params.DefaultWeightMsgRemoveLiquidity
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightAdd,
			SimulateMsgAddLiquidity(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightSwap,
			SimulateMsgSwapOrder(k, ak, bk),
		),

		simulation.NewWeightedOperation(
			weightRemove,
			SimulateMsgRemoveLiquidity(k, ak, bk),
		),
	}
}

// SimulateMsgAddLiquidity  simulates  the addition of liquidity
func SimulateMsgAddLiquidity(k keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (
		opMsg simtypes.OperationMsg, fOps []simtypes.FutureOperation, err error,
	) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		err = FundAccount(r, ctx, k, bk, simAccount.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgAddLiquidity, "unable to fund account"), nil, err
		}
		account := ak.GetAccount(ctx, simAccount.Address)

		var (
			maxToken     sdk.Coin
			minLiquidity sdkmath.Int
		)

		standardDenom, _ := k.GetStandardDenom(ctx)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())
		exactStandardAmt, err := simtypes.RandPositiveInt(r, spendable.AmountOf(standardDenom))
		if err != nil {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgAddLiquidity,
				"standardAmount should be positive",
			), nil, nil
		}
		params := k.GetParams(ctx)
		if exactStandardAmt.GTE(params.MaxStandardCoinPerPool) {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgAddLiquidity,
				"standardAmount should be less than MaxStandardCoinPerPool",
			), nil, nil
		}

		maxToken, err = randToken(r, spendable)
		if err != nil {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgAddLiquidity,
				"insufficient funds",
			), nil, nil
		}

		if maxToken.Denom == standardDenom {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgAddLiquidity,
				"tokenDenom should not be standardDenom",
			), nil, nil
		}

		if strings.HasPrefix(maxToken.Denom, types.LptTokenPrefix) {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgAddLiquidity,
				"tokenDenom should not be liquidity token",
			), nil, nil
		}

		if !maxToken.Amount.IsPositive() {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgAddLiquidity,
				"maxToken must is positive",
			), nil, err
		}

		// check maxToken is registered in MaxSwapAmount
		found := func(denom string) bool {
			MaxSwapAmount := params.MaxSwapAmount
			for _, coin := range MaxSwapAmount {
				if coin.Denom == denom {
					return true
				}
			}
			return false
		}(maxToken.Denom)

		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgAddLiquidity,
				"maxToken is not registered in MaxSwapAmount",
			), nil, err
		}

		poolID := types.GetPoolId(maxToken.Denom)
		pool, has := k.GetPool(ctx, poolID)
		if !has {
			poolCreationFee := k.GetParams(ctx).PoolCreationFee
			spendTotal := poolCreationFee.Amount
			if strings.EqualFold(poolCreationFee.Denom, standardDenom) {
				spendTotal = spendTotal.Add(exactStandardAmt)
			}
			if spendable.AmountOf(poolCreationFee.Denom).LT(spendTotal) {
				return simtypes.NoOpMsg(
					types.ModuleName,
					TypeMsgAddLiquidity,
					"insufficient funds",
				), nil, err
			}
			minLiquidity = exactStandardAmt
		} else {
			balances, err := k.GetPoolBalances(ctx, pool.EscrowAddress)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, TypeMsgAddLiquidity, "pool address not found"), nil, err
			}

			standardReserveAmt := balances.AmountOf(standardDenom)
			if !standardReserveAmt.IsPositive() {
				return simtypes.NoOpMsg(types.ModuleName, TypeMsgAddLiquidity, "standardReserveAmt should be positive"), nil, err
			}
			liquidity := bk.GetSupply(ctx, pool.LptDenom).Amount
			minLiquidity = liquidity.Mul(exactStandardAmt).Quo(standardReserveAmt)

			if !maxToken.Amount.Sub(balances.AmountOf(maxToken.Denom).Mul(exactStandardAmt).Quo(standardReserveAmt)).IsPositive() {
				return simtypes.NoOpMsg(types.ModuleName, TypeMsgAddLiquidity, "insufficient funds"), nil, err
			}
		}

		deadline := randDeadline(r)
		msg := types.NewMsgAddLiquidity(
			maxToken,
			exactStandardAmt,
			minLiquidity,
			deadline,
			account.GetAddress().String(),
		)

		var fees sdk.Coins
		coinsTemp, hasNeg := spendable.SafeSub(
			sdk.NewCoins(sdk.NewCoin(standardDenom, exactStandardAmt), maxToken)...)
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coinsTemp)
			if err != nil {
				return simtypes.NoOpMsg(
					types.ModuleName,
					TypeMsgAddLiquidity,
					"unable to generate fees",
				), nil, nil
			}
		}

		txGen := moduletestutil.MakeTestEncodingConfig().TxConfig
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgAddLiquidity,
				"unable to generate mock tx",
			), nil, err
		}

		if _, _, err := app.SimDeliver(txGen.TxEncoder(), tx); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgAddLiquidity, "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}

}

// SimulateMsgSwapOrder  simulates  the swap of order
func SimulateMsgSwapOrder(k keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (
		opMsg simtypes.OperationMsg, fOps []simtypes.FutureOperation, err error,
	) {
		var (
			inputCoin, outputCoin sdk.Coin
			isBuyOrder            bool
		)

		simAccount, _ := simtypes.RandomAcc(r, accs)
		err = FundAccount(r, ctx, k, bk, simAccount.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgSwapOrder, "unable to fund account"), nil, err
		}
		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())
		standardDenom, _ := k.GetStandardDenom(ctx)

		if spendable.IsZero() {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgSwapOrder,
				"spendable is zero",
			), nil, err
		}

		pools := k.GetAllPools(ctx)
		if len(pools) == 0 {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgSwapOrder,
				"no pool found",
			), nil, nil
		}

		pool := pools[r.Intn(len(pools))]

		reservePool, err := k.GetPoolBalancesByLptDenom(ctx, pool.LptDenom)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "inputCoin should exist in the pool"), nil, nil
		}

		standardReserveAmt := reservePool.AmountOf(standardDenom)
		tokenReserveAmt := reservePool.AmountOf(pool.CounterpartyDenom)

		if !standardReserveAmt.IsPositive() || !tokenReserveAmt.IsPositive() {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgSwapOrder,
				"reserve pool should be positive",
			), nil, nil
		}

		// sold coin
		tokenToStandard := randBoolean(r)
		swapLimit := k.GetParams(ctx).MaxSwapAmount.AmountOf(pool.CounterpartyDenom)
		if tokenToStandard {
			inputCoin = sdk.NewCoin(pool.CounterpartyDenom, simtypes.RandomAmount(r, swapLimit))
			outputCoin = sdk.NewCoin(standardDenom, simtypes.RandomAmount(r, swapLimit))
		} else {
			inputCoin = sdk.NewCoin(standardDenom, simtypes.RandomAmount(r, swapLimit))
			outputCoin = sdk.NewCoin(pool.CounterpartyDenom, simtypes.RandomAmount(r, swapLimit))
		}

		isBuyOrder = randBoolean(r)

		if isBuyOrder {
			inputCoin, outputCoin, err = singleSwapBill(inputCoin, outputCoin, ctx, k)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, TypeMsgSwapOrder, err.Error()), nil, nil
			}
			if tokenToStandard && inputCoin.Amount.GTE(swapLimit) {
				return simtypes.NoOpMsg(
					types.ModuleName,
					TypeMsgSwapOrder,
					"inputCoin amount should be less than swapLimit",
				), nil, nil
			}
			if inputCoin.Amount.GTE(spendable.AmountOf(inputCoin.Denom)) {
				return simtypes.NoOpMsg(
					types.ModuleName,
					TypeMsgSwapOrder,
					"insufficient funds",
				), nil, nil
			}
		} else {
			inputCoin, outputCoin, err = singleSwapSellOrder(inputCoin, outputCoin, ctx, k)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, TypeMsgSwapOrder, err.Error()), nil, nil
			}
			if !tokenToStandard && outputCoin.Amount.GTE(swapLimit) {
				return simtypes.NoOpMsg(
					types.ModuleName,
					TypeMsgSwapOrder,
					"outputCoin amount should be less than swapLimit",
				), nil, nil
			}
			if inputCoin.Amount.GTE(spendable.AmountOf(inputCoin.Denom)) {
				return simtypes.NoOpMsg(
					types.ModuleName,
					TypeMsgSwapOrder,
					"insufficient funds",
				), nil, nil
			}
		}

		if !outputCoin.Amount.IsPositive() {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgSwapOrder,
				"outputCoin must is positive",
			), nil, err
		}

		deadline := randDeadline(r)
		msg := types.NewMsgSwapOrder(
			types.Input{
				Address: simAccount.Address.String(),
				Coin:    inputCoin,
			},
			types.Output{
				Address: simAccount.Address.String(),
				Coin:    outputCoin,
			},
			deadline,
			isBuyOrder,
		)

		var fees sdk.Coins
		coinsTemp, hasNeg := spendable.SafeSub(
			sdk.NewCoins(sdk.NewCoin(inputCoin.Denom, inputCoin.Amount))...)
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coinsTemp)
			if err != nil {
				return simtypes.NoOpMsg(
					types.ModuleName,
					TypeMsgSwapOrder,
					"unable to generate fees",
				), nil, nil
			}
		}

		txGen := moduletestutil.MakeTestEncodingConfig().TxConfig
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(
				types.ModuleName,
				TypeMsgSwapOrder,
				"unable to generate mock tx",
			), nil, err
		}

		if _, _, err := app.SimDeliver(txGen.TxEncoder(), tx); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgSwapOrder, "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgRemoveLiquidity  simulates  the removal of liquidity
func SimulateMsgRemoveLiquidity(k keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (
		opMsg simtypes.OperationMsg, fOps []simtypes.FutureOperation, err error,
	) {

		pools := k.GetAllPools(ctx)
		if len(pools) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "no pool found"), nil, nil
		}

		pool := pools[r.Intn(len(pools))]

		simAccount, err := func(accs []simtypes.Account) (simtypes.Account, error) {
			for _, acc := range accs {
				coins := bk.GetAllBalances(ctx, acc.Address)
				for _, coin := range coins {
					if coin.Denom == pool.LptDenom {
						return acc, nil
					}
				}
			}
			return simtypes.Account{}, errors.New("no account has LptCoin")
		}(accs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, err.Error()), nil, nil
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		standardDenom, _ := k.GetStandardDenom(ctx)

		var (
			minToken          sdkmath.Int
			minStandardAmt    sdkmath.Int
			withdrawLiquidity sdk.Coin
		)

		spendable := bk.SpendableCoins(ctx, account.GetAddress())
		if spendable.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "spendable is zero"), nil, err
		}

		reservePool, err := k.GetPoolBalancesByLptDenom(ctx, pool.LptDenom)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "inputCoin should exist in the pool"), nil, nil
		}

		standardReserveAmt := reservePool.AmountOf(standardDenom)
		tokenReserveAmt := reservePool.AmountOf(pool.CounterpartyDenom)

		withdrawLiquidity = sdk.NewCoin(pool.LptDenom, simtypes.RandomAmount(r, spendable.AmountOf(pool.LptDenom)))
		liquidityReserve := bk.GetSupply(ctx, pool.LptDenom).Amount

		if !withdrawLiquidity.IsValid() || !withdrawLiquidity.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "invalid withdrawLiquidity"), nil, nil
		}
		if liquidityReserve.LT(withdrawLiquidity.Amount) {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "insufficient funds"), nil, nil
		}

		minToken = withdrawLiquidity.Amount.Mul(tokenReserveAmt).Quo(liquidityReserve)
		if tokenReserveAmt.LT(minToken) {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "insufficient funds"), nil, nil
		}

		minStandardAmt = withdrawLiquidity.Amount.Mul(standardReserveAmt).Quo(liquidityReserve)
		if standardReserveAmt.LT(minStandardAmt) {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "insufficient funds"), nil, nil
		}

		deadline := randDeadline(r)
		msg := types.NewMsgRemoveLiquidity(
			minToken,
			withdrawLiquidity,
			minStandardAmt,
			deadline,
			account.GetAddress().String(),
		)

		var fees sdk.Coins
		coinsTemp, hasNeg := spendable.SafeSub(sdk.NewCoins(sdk.NewCoin(pool.CounterpartyDenom, minToken), sdk.NewCoin(standardDenom, minStandardAmt))...)
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coinsTemp)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "unable to generate fees"), nil, nil
			}
		}

		txGen := moduletestutil.MakeTestEncodingConfig().TxConfig

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "unable to generate mock tx"), nil, err
		}

		if _, _, err := app.SimDeliver(txGen.TxEncoder(), tx); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRemoveLiquidity, "unable to deliver tx"), nil, nil
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil

	}
}

func FundAccount(r *rand.Rand, ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper, account sdk.AccAddress) error {
	params := k.GetParams(ctx)
	MaxSwapAmount := params.MaxSwapAmount

	for _, coin := range MaxSwapAmount {
		denom := coin.Denom
		randomAmount := simtypes.RandomAmount(r, sdkmath.NewInt(100000000000000))
		err := bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, randomAmount)))
		if err != nil {
			return errors.New("unable to mint coins")
		}
		err = bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, account, sdk.NewCoins(sdk.NewCoin(denom, randomAmount)))
		if err != nil {
			return errors.New("unable to send coins")
		}
	}
	return nil
}

func randToken(r *rand.Rand, spendableCoin sdk.Coins) (sdk.Coin, error) {
	if len(spendableCoin) == 0 {
		return sdk.Coin{}, errors.New("insufficient funds")
	}
	token := spendableCoin[r.Intn(len(spendableCoin))]
	randAmt, err := simtypes.RandPositiveInt(r, token.Amount.QuoRaw(4))
	if err != nil {
		return sdk.Coin{}, errors.New("insufficient funds")
	}
	return sdk.NewCoin(token.Denom, randAmt), nil
}

func RandomSpendableToken(r *rand.Rand, spendableCoin sdk.Coins) sdk.Coin {
	token := spendableCoin[r.Intn(len(spendableCoin))]
	return sdk.NewCoin(token.Denom, simtypes.RandomAmount(r, token.Amount.QuoRaw(2)))
}

func RandomTotalToken(r *rand.Rand, coins sdk.Coins) sdk.Coin {
	token := coins[r.Intn(len(coins))]
	return sdk.NewCoin(token.Denom, simtypes.RandomAmount(r, token.Amount))
}

func randDeadline(r *rand.Rand) int64 {
	var delta = time.Duration(simtypes.RandIntBetween(r, 10, 100)) * time.Second
	return time.Now().Add(delta).UnixNano()
}

func randBoolean(r *rand.Rand) bool {
	return r.Int()%2 == 0
}

// A single swap bill
func singleSwapBill(inputCoin, outputCoin sdk.Coin, ctx sdk.Context, k keeper.Keeper) (sdk.Coin, sdk.Coin, error) {
	param := k.GetParams(ctx)

	lptDenom, _ := k.GetLptDenomFromDenoms(ctx, outputCoin.Denom, inputCoin.Denom)
	reservePool, _ := k.GetPoolBalancesByLptDenom(ctx, lptDenom)
	outputReserve := reservePool.AmountOf(outputCoin.Denom)
	inputReserve := reservePool.AmountOf(inputCoin.Denom)
	soldTokenAmt := keeper.GetOutputPrice(outputCoin.Amount, inputReserve, outputReserve, param.Fee)

	if soldTokenAmt.IsNegative() {
		return sdk.Coin{}, sdk.Coin{}, errors.New("wrong token price calcualtion")
	}
	inputCoin = sdk.NewCoin(inputCoin.Denom, soldTokenAmt)

	return inputCoin, outputCoin, nil
}

// A single swap sell order
func singleSwapSellOrder(inputCoin, outputCoin sdk.Coin, ctx sdk.Context, k keeper.Keeper) (sdk.Coin, sdk.Coin, error) {
	param := k.GetParams(ctx)

	lptDenom, _ := k.GetLptDenomFromDenoms(ctx, inputCoin.Denom, outputCoin.Denom)
	reservePool, _ := k.GetPoolBalancesByLptDenom(ctx, lptDenom)
	inputReserve := reservePool.AmountOf(inputCoin.Denom)
	outputReserve := reservePool.AmountOf(outputCoin.Denom)
	boughtTokenAmt := keeper.GetInputPrice(inputCoin.Amount, inputReserve, outputReserve, param.Fee)

	outputCoin = sdk.NewCoin(outputCoin.Denom, boughtTokenAmt)
	return inputCoin, outputCoin, nil
}
