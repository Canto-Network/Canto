package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var DefaultLiquidBondDenom = "lscanto"
var RewardPool = DeriveAddress(ModuleName, "RewardPool")

func NativeTokenToLiquidStakeToken(nativeTokenAmount, lsTokenTotalSupplyAmount sdk.Int, netAmount sdk.Dec) (lsTokenAmount sdk.Int) {
	return lsTokenTotalSupplyAmount.ToDec().
		QuoTruncate(netAmount.TruncateDec()).
		MulTruncate(nativeTokenAmount.ToDec()).
		TruncateInt()
}
