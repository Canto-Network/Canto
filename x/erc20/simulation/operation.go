package simulation

import (
	"math/big"
	"math/rand"

	sdkmath "cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v8/contracts"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	"github.com/Canto-Network/Canto/v8/app/params"
	"github.com/Canto-Network/Canto/v8/x/erc20/keeper"
	"github.com/Canto-Network/Canto/v8/x/erc20/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgConvertCoin  = "op_weight_msg_convert_coin"
	OpWeightMsgConvertErc20 = "op_weight_msg_convert_erc20"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	k keeper.Keeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ek types.EVMKeeper,
	fk types.FeeMarketKeeper,
) simulation.WeightedOperations {
	var weightMsgConvertCoinNativeCoin int
	appParams.GetOrGenerate(OpWeightMsgConvertCoin, &weightMsgConvertCoinNativeCoin, nil, func(_ *rand.Rand) {
		weightMsgConvertCoinNativeCoin = params.DefaultWeightMsgConvertCoin
	})

	var weightMsgConvertErc20NativeCoin int
	appParams.GetOrGenerate(OpWeightMsgConvertErc20, &weightMsgConvertErc20NativeCoin, nil, func(_ *rand.Rand) {
		weightMsgConvertErc20NativeCoin = params.DefaultWeightMsgConvertErc20
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgConvertCoinNativeCoin,
			SimulateMsgConvertCoin(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgConvertErc20NativeCoin,
			SimulateMsgConvertErc20(k, ak, bk, ek, fk),
		),
	}
}

// SimulateMsgConvertCoin generates a MsgConvertCoin with random values for convertCoinNativeCoin
func SimulateMsgConvertCoin(k keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		pairs := k.GetTokenPairs(ctx)

		if len(pairs) == 0 {
			_, err := SimulateRegisterCoin(r, ctx, accs, k, bk)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgConvertCoin{}), "no pairs available"), nil, nil
			}
			pairs = k.GetTokenPairs(ctx)
		}

		// randomly pick one pair
		pair := pairs[r.Intn(len(pairs))]
		if !pair.Enabled {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgConvertCoin{}), "token pair is not enabled"), nil, nil
		}
		baseDenom := pair.GetDenom()

		// select random account that has coins baseDenom
		var simAccount simtypes.Account
		var spendable sdk.Coins
		skip := true

		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			spendable = bk.SpendableCoins(ctx, simAccount.Address)
			if spendable.AmountOf(baseDenom).IsPositive() {
				skip = false
				break
			}
		}

		if skip {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgConvertCoin{}), "no account has coins"), nil, nil
		}

		priv, _ := ethsecp256k1.GenerateKey()
		address := common.BytesToAddress(priv.PubKey().Address().Bytes())

		msg := types.NewMsgConvertCoin(
			sdk.NewCoin(baseDenom, spendable.AmountOf(baseDenom)),
			address,
			simAccount.Address,
		)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			CoinsSpentInMsg: spendable,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}
		op, fOps, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
		return op, fOps, err
	}
}

// SimulateMsgConvertErc20 generates a MsgConvertErc20 with random values for convertERC20NativeCoin.
func SimulateMsgConvertErc20(k keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper, ek types.EVMKeeper, fk types.FeeMarketKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		pairs := k.GetTokenPairs(ctx)

		if len(pairs) == 0 {
			_, err := SimulateRegisterERC20(r, ctx, accs, k, ak, bk, ek, fk)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgConvertERC20{}), "no pairs available"), nil, nil
			}
			pairs = k.GetTokenPairs(ctx)
		}

		// randomly pick one pair
		pair := pairs[r.Intn(len(pairs))]
		if !pair.Enabled {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgConvertERC20{}), "token pair is not enabled"), nil, nil
		}

		erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
		deployer := types.ModuleAddress
		contractAddr := pair.GetERC20Contract()
		randomIteration := r.Intn(10)
		for i := 0; i < randomIteration; i++ {
			simAccount, _ := simtypes.RandomAcc(r, accs)

			mintAmt := sdkmath.NewInt(1000000000)
			receiver := common.BytesToAddress(simAccount.Address.Bytes())
			before := k.BalanceOf(ctx, erc20ABI, contractAddr, receiver)
			_, err := k.CallEVM(ctx, erc20ABI, deployer, contractAddr, true, "mint", receiver, mintAmt.BigInt())
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgConvertERC20{}), "no account has native ERC20"), nil, nil
			}
			after := k.BalanceOf(ctx, erc20ABI, contractAddr, receiver)
			if after.Cmp(before.Add(before, mintAmt.BigInt())) != 0 {
				return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgConvertERC20{}), "no account has native ERC20"), nil, nil
			}
		}

		// select random account that has coins baseDenom
		var simAccount simtypes.Account
		var erc20Balance *big.Int
		skip := true

		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			erc20Balance = k.BalanceOf(ctx, erc20ABI, pair.GetERC20Contract(), common.BytesToAddress(simAccount.Address.Bytes()))
			if erc20Balance.Cmp(big.NewInt(0)) > 0 {
				skip = false
				break
			}
		}

		if skip {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgConvertERC20{}), "no account has native ERC20"), nil, nil
		}

		msg := types.NewMsgConvertERC20(sdkmath.NewIntFromBigInt(erc20Balance), simAccount.Address, pair.GetERC20Contract(), common.BytesToAddress(simAccount.Address.Bytes()))

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			CoinsSpentInMsg: sdk.NewCoins(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}

		op, fOps, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
		return op, fOps, err
	}
}
