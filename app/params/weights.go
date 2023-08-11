package params

// Default simulation operation weights for messages and gov proposals.
const (
	DefaultWeightMsgLiquidStake                 int = 50
	DefaultWeightMsgLiquidUnstake               int = 40
	DefaultWeightMsgProvideInsurance            int = 70
	DefaultWeightMsgCancelProvideInsurance      int = 10
	DefaultWeightMsgDepositInsurance            int = 10
	DefaultWeightMsgWithdrawInsurance           int = 20
	DefaultWeightMsgWithdrawInsuranceCommission int = 10
	DefaultWeightMsgClaimDiscountedReward       int = 20
)
