package simulation

import (
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"math/rand"

	"github.com/Canto-Network/Canto/v7/app/params"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/keeper"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
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
		weightMsgLiquidStake = params.DefaultWeightMsgLiquidStake
	})

	var weightMsgLiquidUnstake int
	appParams.GetOrGenerate(cdc, OpWeightMsgLiquidUnstake, &weightMsgLiquidUnstake, nil, func(_ *rand.Rand) {
		weightMsgLiquidUnstake = params.DefaultWeightMsgLiquidUnstake
	})

	var weightMsgProvideInsurance int
	appParams.GetOrGenerate(cdc, OpWeightMsgProvideInsurance, &weightMsgProvideInsurance, nil, func(_ *rand.Rand) {
		weightMsgProvideInsurance = params.DefaultWeightMsgProvideInsurance
	})

	var weightMsgCancelProvideInsurance int
	appParams.GetOrGenerate(cdc, OpWeightMsgCancelProvideInsurance, &weightMsgCancelProvideInsurance, nil, func(_ *rand.Rand) {
		weightMsgCancelProvideInsurance = params.DefaultWeightMsgCancelProvideInsurance
	})

	var weightMsgDepositInsurance int
	appParams.GetOrGenerate(cdc, OpWeightMsgDepositInsurance, &weightMsgDepositInsurance, nil, func(_ *rand.Rand) {
		weightMsgDepositInsurance = params.DefaultWeightMsgDepositInsurance
	})

	var weightMsgWithdrawInsurance int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawInsurance, &weightMsgWithdrawInsurance, nil, func(_ *rand.Rand) {
		weightMsgWithdrawInsurance = params.DefaultWeightMsgWithdrawInsurance
	})

	var weightMsgWithdrawInsuranceCommission int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawInsuranceCommission, &weightMsgWithdrawInsuranceCommission, nil, func(_ *rand.Rand) {
		weightMsgWithdrawInsuranceCommission = params.DefaultWeightMsgWithdrawInsuranceCommission
	})

	var weightMsgClaimDiscountedReward int
	appParams.GetOrGenerate(cdc, OpWeightMsgClaimDiscountedReward, &weightMsgClaimDiscountedReward, nil, func(_ *rand.Rand) {
		weightMsgClaimDiscountedReward = params.DefaultWeightMsgClaimDiscountedReward
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLiquidStake,
			SimulateMsgLiquidStake(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgLiquidUnstake,
			SimulateMsgLiquidUnstake(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgProvideInsurance,
			SimulateMsgProvideInsurance(ak, bk, sk),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelProvideInsurance,
			SimulateMsgCancelProvideInsurance(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgDepositInsurance,
			SimulateMsgDepositInsurance(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawInsurance,
			SimulateMsgWithdrawInsurance(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawInsuranceCommission,
			SimulateMsgWithdrawInsuranceCommission(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgClaimDiscountedReward,
			SimulateMsgClaimDiscountedReward(ak, bk, k),
		),
	}
}

// SimulateMsgLiquidStake generates a MsgLiquidStake with random values.
func SimulateMsgLiquidStake(ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		bondDenom := sk.BondDenom(ctx)
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)
		delegator := account.GetAddress()
		spendable := bk.SpendableCoins(ctx, delegator)

		chunksToLiquidStake := int64(simtypes.RandIntBetween(r, 1, 3))
		nas := k.GetNetAmountState(ctx)
		lsmParams := k.GetParams(ctx)
		totalSupplyAmt := bk.GetSupply(ctx, bondDenom).Amount
		availableChunkSlots := types.GetAvailableChunkSlots(nas.UtilizationRatio, lsmParams.DynamicFeeRate.UHardCap, totalSupplyAmt).Int64()
		if availableChunkSlots == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidStake, "no available chunk slots"), nil, nil
		}

		pairingInsurances, _ := k.GetPairingInsurances(ctx)
		if len(pairingInsurances) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidStake, "no pairing insurances"), nil, nil
		}
		if chunksToLiquidStake > availableChunkSlots {
			chunksToLiquidStake = availableChunkSlots
		}
		if len(pairingInsurances) < int(chunksToLiquidStake) {
			chunksToLiquidStake = int64(len(pairingInsurances))
		}

		stakingCoins := sdk.NewCoins(
			sdk.NewCoin(
				bondDenom,
				types.ChunkSize.MulRaw(chunksToLiquidStake),
			),
		)
		if !spendable.AmountOf(bondDenom).GTE(stakingCoins[0].Amount) {
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
			CoinsSpentInMsg: stakingCoins,
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
func SimulateMsgLiquidUnstake(ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		bondDenom := sk.BondDenom(ctx)
		var simAccount simtypes.Account
		var delegator sdk.AccAddress
		var spendable sdk.Coins

		liquidBondDenom := k.GetLiquidBondDenom(ctx)
		nas := k.GetNetAmountState(ctx)
		if nas.MintRate.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidUnstake, "cannot unstake because there are no chunks"), nil, nil
		}
		lsTokensToPayForOneChunk := nas.MintRate.Mul(types.ChunkSize.ToDec()).Ceil().TruncateInt()
		for i := 0; i < len(accs); i++ {
			simAccount, _ = simtypes.RandomAcc(r, accs)
			account := ak.GetAccount(ctx, simAccount.Address)
			spendable = bk.SpendableCoins(ctx, simAccount.Address)

			delegator = account.GetAddress()
			// delegator must have enough ls tokens to pay for one chunk
			if spendable.AmountOf(liquidBondDenom).GTE(lsTokensToPayForOneChunk) {
				break
			}
		}
		if !spendable.AmountOf(liquidBondDenom).GTE(lsTokensToPayForOneChunk) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidUnstake, "not enough ls tokens to liquid unstake an one chunk"), nil, nil
		}

		maxAvailableNumChunksToLiquidUnstake := spendable.AmountOf(types.DefaultLiquidBondDenom).Quo(lsTokensToPayForOneChunk).Int64()
		if maxAvailableNumChunksToLiquidUnstake == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidUnstake, "not enough ls tokens to liquid unstake an one chunk"), nil, nil
		}

		var chunksToLiquidUnstake int64
		if maxAvailableNumChunksToLiquidUnstake > 1 {
			chunksToLiquidUnstake = int64(simtypes.RandIntBetween(r, 1, int(maxAvailableNumChunksToLiquidUnstake)))
		} else {
			chunksToLiquidUnstake = maxAvailableNumChunksToLiquidUnstake
		}
		// delegator can liquid unstake one or more chunks
		var pairedChunks []types.Chunk
		k.IterateAllChunks(ctx, func(chunk types.Chunk) bool {
			if chunk.Status == types.CHUNK_STATUS_PAIRED {
				// check whether the chunk is already have unstaking requests in queue.
				_, found := k.GetUnpairingForUnstakingChunkInfo(ctx, chunk.Id)
				if found {
					return false
				}
				pairedChunks = append(pairedChunks, chunk)
			}
			return false
		})

		numPairedChunks := int64(len(pairedChunks))
		if numPairedChunks == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgLiquidUnstake, "no paired chunks"), nil, nil
		}

		if numPairedChunks < maxAvailableNumChunksToLiquidUnstake {
			chunksToLiquidUnstake = numPairedChunks
		}

		unstakingCoin := sdk.NewCoin(bondDenom, types.ChunkSize.MulRaw(chunksToLiquidUnstake))

		msg := types.NewMsgLiquidUnstake(delegator.String(), unstakingCoin)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simapp.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: sdk.NewCoins(unstakingCoin),
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
		bondDenom := sk.BondDenom(ctx)
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)
		provider := account.GetAddress()
		spendable := bk.SpendableCoins(ctx, provider)

		upperThanMinimumCollateral := simtypes.RandomDecAmount(r, sdk.MustNewDecFromStr("0.03"))
		minCollateral := sdk.MustNewDecFromStr(types.MinimumCollateral)
		minCollateral = minCollateral.Add(upperThanMinimumCollateral)
		collaterals := sdk.NewCoins(
			sdk.NewCoin(
				bondDenom,
				minCollateral.Mul(types.ChunkSize.ToDec()).Ceil().TruncateInt(),
			),
		)
		maximumFeeRate := sdk.MustNewDecFromStr("0.5")
		var validators []stakingtypes.Validator
		sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
			// Only select validators with commission rate less than 50%
			if validator.GetCommission().LT(maximumFeeRate) {
				v, ok := validator.(stakingtypes.Validator)
				if !ok {
					return false
				}
				validators = append(validators, v)
			}
			return false
		})
		if len(validators) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgProvideInsurance, "no validators to provide insurance"), nil, nil
		}

		// select one validator randomly
		validator := validators[r.Intn(len(validators))]

		feeRate := simtypes.RandomDecAmount(r, sdk.MustNewDecFromStr("0.15"))
		if validator.GetCommission().Add(feeRate).GTE(maximumFeeRate) {
			feeRate = maximumFeeRate.Sub(validator.GetCommission()).Sub(sdk.MustNewDecFromStr("0.001"))
		}

		if !spendable.AmountOf(bondDenom).GTE(collaterals[0].Amount) {
			if err := bk.MintCoins(ctx, types.ModuleName, collaterals); err != nil {
				panic(err)
			}
			if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, provider, collaterals); err != nil {
				panic(err)
			}
			spendable = bk.SpendableCoins(ctx, provider)
		}

		if feeRate.IsNegative() {
			feeRate = sdk.ZeroDec()
		}
		msg := types.NewMsgProvideInsurance(provider.String(), validator.GetOperator().String(), collaterals[0], feeRate)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simapp.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: collaterals,
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
func SimulateMsgCancelProvideInsurance(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
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
				if insurance.GetProvider().Equals(provider) && insurance.Status == types.INSURANCE_STATUS_PAIRING {
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
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}
		return types.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgDepositInsurance generates a MsgDepositInsurance with random values.
func SimulateMsgDepositInsurance(ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		bondDenom := sk.BondDenom(ctx)
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
			bondDenom,
			minCollateral.Mul(types.ChunkSize.ToDec()).Ceil().TruncateInt(),
		)

		// deposit 1 % ~ 10 % of the collateral
		depositPortion := types.RandomDec(r, sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.1"))
		deposits := sdk.NewCoins(
			sdk.NewCoin(
				bondDenom,
				collateral.Amount.ToDec().Mul(depositPortion).TruncateInt(),
			),
		)

		if !spendable.AmountOf(bondDenom).GTE(deposits[0].Amount) {
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
			CoinsSpentInMsg: deposits,
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
func SimulateMsgWithdrawInsurance(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
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
					if insurance.Status == types.INSURANCE_STATUS_PAIRED || insurance.Status == types.INSURANCE_STATUS_UNPAIRED {
						withdrawableInsurances = append(withdrawableInsurances, insurance)
					}
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
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdrawInsurance, "no withdrawable ins"), nil, nil
		}
		// select randomly one ins to withdraw
		ins := withdrawableInsurances[r.Intn(len(withdrawableInsurances))]
		insBals := bk.SpendableCoins(ctx, ins.DerivedAddress())
		if !insBals.IsValid() || !insBals.IsAllPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdrawInsurance, "no withdrawable insurance coins"), nil, nil
		}
		msg := types.NewMsgWithdrawInsurance(ins.GetProvider().String(), ins.Id)
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
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}
		return types.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgWithdrawInsuranceCommission generates a MsgWithdrawInsuranceCommission with random values.
func SimulateMsgWithdrawInsuranceCommission(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
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
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdrawInsuranceCommission, "no withdrawable ins"), nil, nil
		}
		// select randomly one ins to withdraw
		ins := withdrawableInsurances[r.Intn(len(withdrawableInsurances))]
		feePoolBals := bk.SpendableCoins(ctx, ins.FeePoolAddress())
		if !feePoolBals.IsValid() || !feePoolBals.IsAllPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWithdrawInsuranceCommission, "no withdrawable fee pool coins"), nil, nil
		}
		msg := types.NewMsgWithdrawInsuranceCommission(ins.GetProvider().String(), ins.Id)
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
			Bankkeeper:      bk,
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

		liquidBondDenom := k.GetLiquidBondDenom(ctx)

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
			if spendable.AmountOf(liquidBondDenom).IsPositive() {
				break
			}
		}
		maxLsTokensToGetAllRewards := nas.MintRate.Mul(minimumDiscountRate).Mul(nas.RewardModuleAccBalance.ToDec()).Ceil().TruncateInt()
		amountToUse := types.RandomInt(r, spendable.AmountOf(liquidBondDenom), maxLsTokensToGetAllRewards)
		lsTokensToUse := sdk.NewCoins(sdk.NewCoin(liquidBondDenom, amountToUse))

		msg := types.NewMsgClaimDiscountedReward(lsTokenHolder.String(), sdk.NewCoin(liquidBondDenom, amountToUse), minimumDiscountRate)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simapp.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: lsTokensToUse,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}
		return types.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}
