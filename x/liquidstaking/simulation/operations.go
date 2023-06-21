package simulation

import (
	"math/rand"

	"github.com/Canto-Network/Canto/v6/app"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
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
func WeightedOperations(appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simulation.WeightedOperations {
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
		// simulation.NewWeightedOperation(
		// 	weightMsgLiquidStake,
		// 	SimulateMsgLiquidStake(ak, bk, k),
		// ),
	}
}

// TODO: Implement simulation
// func SimulateMsgLiquidStake(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
// 	return func(
// 		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
// 	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
// 		simAccount, _ := simtypes.RandomAcc(r, accs)
// 		account := ak.GetAccount(ctx, simAccount.Address)
// 		delegator := account.GetAddress()
// 	}
// }
