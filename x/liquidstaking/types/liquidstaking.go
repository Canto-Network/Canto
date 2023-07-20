package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const Empty uint64 = 0

// SecurityCap is a maximum cap of utilization ratio in module.
// min(UHardcap, SecurityCap) is used when check available chunk slots.
var SecurityCap = sdk.MustNewDecFromStr("0.25")
var MaximumDiscountRate = sdk.MustNewDecFromStr("0.03")

// MaximumInsuranceFeeRate is a maximum cap of insurance + validator fee rate.
var MaximumInsValFeeRate = sdk.MustNewDecFromStr("0.5")

var DefaultLiquidBondDenom = "lscanto"
var RewardPool = DeriveAddress(ModuleName, "RewardPool")
var LsTokenEscrowAcc = DeriveAddress(ModuleName, "LsTokenEscrowAcc")

// NativeTokenToLiquidStakeToken calculate ls token amount from native token amount.
// return (ls token total supply / net amount * native token amount)
func NativeTokenToLiquidStakeToken(
	nativeTokenAmount, lsTokenTotalSupplyAmount sdk.Int,
	netAmount sdk.Dec,
) (lsTokenAmount sdk.Int) {
	return lsTokenTotalSupplyAmount.ToDec().
		QuoTruncate(netAmount.TruncateDec()).
		MulTruncate(nativeTokenAmount.ToDec()).
		TruncateInt()
}
