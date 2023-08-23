package simulation

import (
	"math/big"
	"math/rand"

	"github.com/Canto-Network/Canto/v7/contracts"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	"github.com/Canto-Network/Canto/v7/app/params"
	"github.com/Canto-Network/Canto/v7/x/erc20/keeper"
	"github.com/Canto-Network/Canto/v7/x/erc20/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgConvertCoin  = "op_weight_msg_convert_coin"
	OpWeightMsgConvertErc20 = "op_weight_msg_convert_erc20"
)

var (
	Gas  = uint64(20000000)
	Fees = sdk.Coins{
		{
			Denom:  sdk.DefaultBondDenom,
			Amount: sdk.NewInt(0),
		},
	}
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgConvertCoinNativeCoin int
	appParams.GetOrGenerate(cdc, OpWeightMsgConvertCoin, &weightMsgConvertCoinNativeCoin, nil, func(_ *rand.Rand) {
		weightMsgConvertCoinNativeCoin = params.DefaultWeightMsgConvertCoin
	})

	var weightMsgConvertErc20NativeCoin int
	appParams.GetOrGenerate(cdc, OpWeightMsgConvertErc20, &weightMsgConvertErc20NativeCoin, nil, func(_ *rand.Rand) {
		weightMsgConvertErc20NativeCoin = params.DefaultWeightMsgConvertErc20
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgConvertCoinNativeCoin,
			SimulateMsgConvertCoin(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgConvertErc20NativeCoin,
			SimulateMsgConvertErc20(ak, bk, k),
		),
	}
}

// SimulateMsgConvertCoin generates a MsgConvertCoin with random values for convertCoinNativeCoin
func SimulateMsgConvertCoin(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		pairs := k.GetTokenPairs(ctx)

		if len(pairs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgConvertCoin, "no pairs available"), nil, nil
		}

		// randomly pick one pair
		pair := pairs[r.Intn(len(pairs))]
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
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgConvertCoin, "no account has coins"), nil, nil
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
			TxGen:           simapp.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: spendable,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}
		return types.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgConvertErc20 generates a MsgConvertErc20 with random values for convertERC20NativeCoin.
func SimulateMsgConvertErc20(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		pairs := k.GetTokenPairs(ctx)

		if len(pairs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgConvertERC20, "no pairs available"), nil, nil
		}

		// randomly pick one pair
		pair := pairs[r.Intn(len(pairs))]

		// select random account that has coins baseDenom
		var simAccount simtypes.Account
		var erc20Balance *big.Int
		erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
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
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgConvertERC20, "no account has native ERC20"), nil, nil
		}

		msg := types.NewMsgConvertERC20(sdk.NewIntFromBigInt(erc20Balance), simAccount.Address, pair.GetERC20Contract(), common.BytesToAddress(simAccount.Address.Bytes()))

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simapp.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: sdk.NewCoins(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}
		return types.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}
