package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v6/app"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgLiquidStake                 = "op_weight_msg_liquid_stake"
	OpWeightMsgLiquidUnstake               = "op_weight_msg_liquid_unstake"
	OpWeightMsgProvideInsurance            = "op_weight_msg_provide_insurance"
	OpWeightMsgCancelProvideInsurance      = "op_weight_msg_cancel_provide_insurance"
	OpWeightMsgDepositInsurance            = "op_weight_msg_deposit_insurance"
	OpWeightMsgWithdrawInsurance           = "op_weight_msg_withdraw_insurance"
	OpWeightMsgWithdrawInsuranceCommission = "op_weight_msg_withdraw_insurance_commission"
	OpWeightMsgClaimDiscountedReward       = "op_weight_msg_claim_discounted_reward"
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
	sk types.StakingKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgLiquidStake int
	appParams.GetOrGenerate(cdc, OpWeightMsgLiquidStake, &weightMsgLiquidStake, nil, func(_ *rand.Rand) {
		weightMsgLiquidStake = app.DefaultWeightMsgLiquidStake
	})

	var weightMsgLiquidUnstake int
	appParams.GetOrGenerate(cdc, OpWeightMsgLiquidUnstake, &weightMsgLiquidUnstake, nil, func(_ *rand.Rand) {
		weightMsgLiquidUnstake = app.DefaultWeightMsgLiquidUnstake
	})

	var weightMsgProvideInsurance int
	appParams.GetOrGenerate(cdc, OpWeightMsgProvideInsurance, &weightMsgProvideInsurance, nil, func(_ *rand.Rand) {
		weightMsgProvideInsurance = app.DefaultWeightMsgProvideInsurance
	})

	var weightMsgCancelProvideInsurance int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelProvideInsurance, &weightMsgCancelProvideInsurance, nil, func(_ *rand.Rand) {
		weightMsgCancelProvideInsurance = app.DefaultWeightMsgCancelProvideInsurance
	})

	var weightMsgDepositInsurance int
	appParams.GetOrGenerate(cdc, OpWeightMsgDepositInsurance, &weightMsgDepositInsurance, nil, func(_ *rand.Rand) {
		weightMsgDepositInsurance = app.DefaultWeightMsgDepositInsurance
	})

	var weightMsgWithdrawInsurance int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawInsurance, &weightMsgWithdrawInsurance, nil, func(_ *rand.Rand) {
		weightMsgWithdrawInsurance = app.DefaultWeightMsgWithdrawInsurance
	})

	var weightMsgWithdrawInsuranceCommission int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawInsuranceCommission, &weightMsgWithdrawInsuranceCommission, nil, func(_ *rand.Rand) {
		weightMsgWithdrawInsuranceCommission = app.DefaultWeightMsgWithdrawInsuranceCommission
	})

	var weightMsgClaimDiscountedReward int
	appParams.GetOrGenerate(cdc, OpWeightMsgClaimDiscountedReward, &weightMsgClaimDiscountedReward, nil, func(_ *rand.Rand) {
		weightMsgClaimDiscountedReward = app.DefaultWeightMsgClaimDiscountedReward
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLiquidStake,
			SimulateMsgLiquidStake(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgLiquidUnstake,
			SimulateMsgLiquidUnstake(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgProvideInsurance,
			SimulateMsgProvideInsurance(ak, bk, sk),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelProvideInsurance,
			SimulateMsgCancelProvideInsurance(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgDepositInsurance,
			SimulateMsgDepositInsurance(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawInsurance,
			SimulateMsgWithdrawInsurance(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawInsuranceCommission,
			SimulateMsgWithdrawInsuranceCommission(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgClaimDiscountedReward,
			SimulateMsgClaimDiscountedReward(ak, bk, k),
		),
	}
}

// TODO: add msgs for staking module

// SimulateMsgLiquidStake generates a MsgLiquidStake with random values.
func SimulateMsgLiquidStake(ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)
		delegator := account.GetAddress()
		spendable := bk.SpendableCoins(ctx, delegator)

		chunksToLiquidStake := int64(simtypes.RandIntBetween(r, 1, 3))
		stakingCoins := sdk.NewCoins(
			sdk.NewCoin(
				sdk.DefaultBondDenom,
				types.ChunkSize.MulRaw(chunksToLiquidStake),
			),
		)
		if !spendable.AmountOf(sdk.DefaultBondDenom).GTE(stakingCoins[0].Amount) {
			if err := bk.MintCoins(ctx, types.ModuleName, stakingCoins); err != nil {
				panic(err)
			}
			if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegator, stakingCoins); err != nil {
				panic(err)
			}
			spendable = bk.SpendableCoins(ctx, delegator)
		}

		msg := types.NewMsgLiquidStake(delegator.String(), stakingCoins[0])
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

// SimulateMsgLiquidUnstake generates a MsgLiquidUnstake with random values.
func SimulateMsgLiquidUnstake(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account
		var delegator sdk.AccAddress
		var spendable sdk.Coins

		nas := k.GetNetAmountState(ctx)
		if nas.MintRate.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidUnstake, "cannot unstake because there are no chunks"), nil, nil
		}
		oneChunk, _ := k.GetMinimumRequirements(ctx)
		lsTokensToPayForOneChunk := nas.MintRate.Mul(oneChunk.Amount.ToDec()).Ceil().TruncateInt()
		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			account := ak.GetAccount(ctx, simAccount.Address)
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			delegator = account.GetAddress()
			// delegator must have enough ls tokens to pay for one chunk
			if spendable.AmountOf(types.DefaultLiquidBondDenom).GTE(lsTokensToPayForOneChunk) {
				break
			}
		}
		if !spendable.AmountOf(types.DefaultLiquidBondDenom).GTE(lsTokensToPayForOneChunk) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidUnstake, "not enough ls tokens to liquid unstake an one chunk"), nil, nil
		}

		maxAvailableNumChunksToLiquidUnstake := spendable.AmountOf(types.DefaultLiquidBondDenom).Quo(lsTokensToPayForOneChunk)

		// delegator can liquid unstake one or more chunks
		chunksToLiquidStake := int64(simtypes.RandIntBetween(r, 1, int(maxAvailableNumChunksToLiquidUnstake.Int64())))
		unstakingCoin := sdk.NewCoin(sdk.DefaultBondDenom, oneChunk.Amount.MulRaw(chunksToLiquidStake))

		msg := types.NewMsgLiquidUnstake(delegator.String(), unstakingCoin)
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

// SimulateMsgProvideInsurance generates a MsgProvideInsurance with random values.
func SimulateMsgProvideInsurance(ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)
		provider := account.GetAddress()
		spendable := bk.SpendableCoins(ctx, provider)

		upperThanMinimumCollateral := simtypes.RandomDecAmount(r, sdk.MustNewDecFromStr("0.03"))
		minCollateral := sdk.MustNewDecFromStr(types.MinimumCollateral)
		minCollateral = minCollateral.Add(upperThanMinimumCollateral)
		collaterals := sdk.NewCoins(
			sdk.NewCoin(
				sdk.DefaultBondDenom,
				minCollateral.Ceil().TruncateInt(),
			),
		)
		feeRate := simtypes.RandomDecAmount(r, sdk.MustNewDecFromStr("0.15"))

		if !spendable.AmountOf(sdk.DefaultBondDenom).GTE(collaterals[0].Amount) {
			if err := bk.MintCoins(ctx, types.ModuleName, collaterals); err != nil {
				panic(err)
			}
			if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, provider, collaterals); err != nil {
				panic(err)
			}
			spendable = bk.SpendableCoins(ctx, provider)
		}

		validators := sk.GetAllValidators(ctx)
		// select one validator randomly
		validator := validators[r.Intn(len(validators))]

		msg := types.NewMsgProvideInsurance(provider.String(), validator.GetOperator().String(), collaterals[0], feeRate)
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

// SimulateMsgCancelProvideInsurance generates a MsgCancelProvideInsurance with random values.
func SimulateMsgCancelProvideInsurance(ak types.AccountKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account
		var provider sdk.AccAddress

		cancelableInsurances := make([]types.Insurance, 0)
		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			account := ak.GetAccount(ctx, simAccount.Address)
			provider = account.GetAddress()
			k.IterateAllInsurances(ctx, func(insurance types.Insurance) bool {
				if insurance.GetProvider().Equals(provider) {
					cancelableInsurances = append(cancelableInsurances, insurance)
				}
				return false
			})
			if len(cancelableInsurances) == 0 {
				// Initiate a new insurances
				cancelableInsurances = cancelableInsurances[:0]
				continue
			} else {
				break
			}
		}
		if len(cancelableInsurances) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCancelProvideInsurance, "no cancelable insurance"), nil, nil
		}
		// select randomly one insurance to cancel
		insurance := cancelableInsurances[r.Intn(len(cancelableInsurances))]
		msg := types.NewMsgCancelProvideInsurance(insurance.GetProvider().String(), insurance.Id)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simapp.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: nil,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      nil,
			ModuleName:      types.ModuleName,
		}
		return types.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgDepositInsurance generates a MsgDepositInsurance with random values.
func SimulateMsgDepositInsurance(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account
		var provider sdk.AccAddress
		var spendable sdk.Coins

		depositableInsurances := make([]types.Insurance, 0)
		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			account := ak.GetAccount(ctx, simAccount.Address)
			provider = account.GetAddress()
			spendable = bk.SpendableCoins(ctx, provider)

			k.IterateAllInsurances(ctx, func(insurance types.Insurance) bool {
				if insurance.GetProvider().Equals(provider) {
					depositableInsurances = append(depositableInsurances, insurance)
				}
				return false
			})
			if len(depositableInsurances) == 0 {
				// Initiate a new insurances
				depositableInsurances = depositableInsurances[:0]
				continue
			} else {
				break
			}
		}
		if len(depositableInsurances) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDepositInsurance, "no depositable insurance"), nil, nil
		}
		// select randomly one insurance to cancel
		insurance := depositableInsurances[r.Intn(len(depositableInsurances))]

		minCollateral := sdk.MustNewDecFromStr(types.MinimumCollateral)
		collateral := sdk.NewCoin(
			sdk.DefaultBondDenom,
			minCollateral.Ceil().TruncateInt(),
		)

		// deposit 1 % ~ 10 % of the collateral
		depositPortion := types.RandomDec(r, sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.1"))
		deposits := sdk.NewCoins(
			sdk.NewCoin(
				sdk.DefaultBondDenom,
				collateral.Amount.ToDec().Mul(depositPortion).TruncateInt(),
			),
		)

		if !spendable.AmountOf(sdk.DefaultBondDenom).GTE(deposits[0].Amount) {
			if err := bk.MintCoins(ctx, types.ModuleName, deposits); err != nil {
				panic(err)
			}
			if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, provider, deposits); err != nil {
				panic(err)
			}
			spendable = bk.SpendableCoins(ctx, provider)
		}

		msg := types.NewMsgDepositInsurance(provider.String(), insurance.Id, deposits[0])
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

// SimulateMsgWithdrawInsurance generates a MsgWithdrawInsurance with random values.
func SimulateMsgWithdrawInsurance(ak types.AccountKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account
		var provider sdk.AccAddress

		withdrawableInsurances := make([]types.Insurance, 0)
		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			account := ak.GetAccount(ctx, simAccount.Address)
			provider = account.GetAddress()
			k.IterateAllInsurances(ctx, func(insurance types.Insurance) bool {
				if insurance.GetProvider().Equals(provider) {
					withdrawableInsurances = append(withdrawableInsurances, insurance)
				}
				return false
			})
			if len(withdrawableInsurances) == 0 {
				// Initiate a new insurances
				withdrawableInsurances = withdrawableInsurances[:0]
				continue
			} else {
				break
			}
		}
		if len(withdrawableInsurances) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdrawInsurance, "no withdrawable insurance"), nil, nil
		}
		// select randomly one insurance to withdraw
		insurance := withdrawableInsurances[r.Intn(len(withdrawableInsurances))]
		msg := types.NewMsgWithdrawInsurance(insurance.GetProvider().String(), insurance.Id)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simapp.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: nil,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      nil,
			ModuleName:      types.ModuleName,
		}
		return types.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgWithdrawInsuranceCommission generates a MsgWithdrawInsuranceCommission with random values.
func SimulateMsgWithdrawInsuranceCommission(ak types.AccountKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account
		var provider sdk.AccAddress

		withdrawableInsurances := make([]types.Insurance, 0)
		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			account := ak.GetAccount(ctx, simAccount.Address)
			provider = account.GetAddress()
			k.IterateAllInsurances(ctx, func(insurance types.Insurance) bool {
				if insurance.GetProvider().Equals(provider) {
					withdrawableInsurances = append(withdrawableInsurances, insurance)
				}
				return false
			})
			if len(withdrawableInsurances) == 0 {
				// Initiate a new insurances
				withdrawableInsurances = withdrawableInsurances[:0]
				continue
			} else {
				break
			}
		}
		if len(withdrawableInsurances) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdrawInsuranceCommission, "no withdrawable insurance"), nil, nil
		}
		// select randomly one insurance to withdraw
		insurance := withdrawableInsurances[r.Intn(len(withdrawableInsurances))]
		msg := types.NewMsgWithdrawInsuranceCommission(insurance.GetProvider().String(), insurance.Id)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simapp.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: nil,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      nil,
			ModuleName:      types.ModuleName,
		}
		return types.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgClaimDiscountedReward generates a MsgClaimDiscountedReward with random values.
func SimulateMsgClaimDiscountedReward(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account
		var lsTokenHolder sdk.AccAddress
		var spendable sdk.Coins

		nas := k.GetNetAmountState(ctx)
		if !nas.DiscountRate.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgClaimDiscountedReward, "discount rate is zero"), nil, nil
		}
		onePercent := sdk.NewDecWithPrec(1, 2)
		// When the discount rate is less than 1%, arbitrager will not claim discounted reward
		minimumDiscountRate := sdk.MinDec(nas.DiscountRate, onePercent)
		if minimumDiscountRate.LT(onePercent) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgClaimDiscountedReward, "discount rate is less than 1%"), nil, nil
		}

		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			account := ak.GetAccount(ctx, simAccount.Address)
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			lsTokenHolder = account.GetAddress()
			// delegator must have enough ls tokens to pay for one chunk
			if spendable.AmountOf(types.DefaultLiquidBondDenom).IsPositive() {
				break
			}
		}
		maxLsTokensToGetAllRewards := nas.MintRate.Mul(minimumDiscountRate).Mul(nas.RewardModuleAccBalance.ToDec()).Ceil().TruncateInt()
		amountToUse := types.RandomInt(r, spendable.AmountOf(types.DefaultLiquidBondDenom), maxLsTokensToGetAllRewards)

		msg := types.NewMsgClaimDiscountedReward(lsTokenHolder.String(), sdk.NewCoin(sdk.DefaultBondDenom, amountToUse), minimumDiscountRate)
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
