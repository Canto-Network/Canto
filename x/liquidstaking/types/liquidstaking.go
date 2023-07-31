package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const Empty uint64 = 0

// SecurityCap is a maximum cap of utilization ratio in module.
// min(UHardcap, SecurityCap) is used when check available chunk slots.
var SecurityCap = sdk.MustNewDecFromStr("0.25")
var MaximumDiscountRateCap = sdk.MustNewDecFromStr("0.1")

// MaximumInsuranceFeeRate is a maximum cap of insurance + validator fee rate.
var MaximumInsValFeeRate = sdk.MustNewDecFromStr("0.5")

var DefaultLiquidBondDenom = "lscanto"
var RewardPool = DeriveAddress(ModuleName, "RewardPool")
var LsTokenEscrowAcc = DeriveAddress(ModuleName, "LsTokenEscrowAcc")
