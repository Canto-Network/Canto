package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const Empty uint64 = 0

var DefaultLiquidBondDenom = "lscanto"
var RewardPool = DeriveAddress(ModuleName, "RewardPool")
var LsTokenEscrowAcc = DeriveAddress(ModuleName, "LsTokenEscrowAcc")

func NativeTokenToLiquidStakeToken(
	nativeTokenAmount, lsTokenTotalSupplyAmount sdk.Int,
	netAmount sdk.Dec,
) (lsTokenAmount sdk.Int) {
	return lsTokenTotalSupplyAmount.ToDec().
		QuoTruncate(netAmount.TruncateDec()).
		MulTruncate(nativeTokenAmount.ToDec()).
		TruncateInt()
}

func LiquidStakeTokenToNativeToken(
	lsTokenAmount, lsTokenTotalSupplyAmount sdk.Int,
	netAmount sdk.Dec,
) (nativeTokenAmount sdk.Int) {
	return lsTokenAmount.ToDec().
		Mul(netAmount).
		Quo(lsTokenTotalSupplyAmount.ToDec()).
		TruncateInt()
}
